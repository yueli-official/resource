import type { ResourceAssetView } from '~/types'
import { assetUploadURL } from '@yueli/asset-nuxt/upload'

// useUpload drives the asset service's three-step upload: the resource
// backend mints a presigned blob URL (init), the browser PUTs the bytes straight
// to the asset service (bypassing the BFF proxy — hence the asset service's CORS),
// then the backend finalizes + links the asset. init/finalize go through the BFF
// (`useApi`, Bearer-injected); the PUT is a direct, credential-less HMAC link.
export function useUpload() {
  const { call } = useApi()
  const multipartThreshold = 64 * 1024 * 1024
  const multipartConcurrency = 3

  type UploadInit = {
    uploadUrl: string
    uploadToken: string
    method?: 'PUT' | 'MULTIPART' | string
    uploadHeaders?: Record<string, string>
    uploadId?: string
    partSize?: number
    partCount?: number
  }
  type MultipartPartURL = { uploadUrl: string; uploadHeaders?: Record<string, string> }
  type CompletedPart = { partNumber: number; etag: string }

  // putWithProgress streams a File to an absolute presigned URL, reporting bytes
  // sent. XHR is used because fetch() can't report upload progress. The asset
  // service derives the stored content-type from the signed token, so the request
  // Content-Type is irrelevant here.
  function putBlob(
    url: string,
    body: Blob,
    onLoaded?: (loaded: number, total: number) => void,
    headers?: Record<string, string>
  ): Promise<XMLHttpRequest> {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest()
      xhr.open('PUT', assetUploadURL(url))
      for (const [key, value] of Object.entries(headers ?? {})) xhr.setRequestHeader(key, value)
      xhr.upload.onprogress = (e) => {
        if (e.lengthComputable) onLoaded?.(e.loaded, e.total)
      }
      xhr.onload = () =>
        xhr.status >= 200 && xhr.status < 300
          ? resolve(xhr)
          : reject(new Error(`上传失败 (HTTP ${xhr.status})`))
      xhr.onerror = () => reject(new Error('上传网络错误(检查素材服务 CORS / 是否在线)'))
      xhr.send(body)
    })
  }

  function putWithProgress(url: string, file: File, onProgress?: (pct: number) => void, headers?: Record<string, string>): Promise<XMLHttpRequest> {
    return putBlob(url, file, (loaded, total) => {
      onProgress?.(Math.round((loaded / total) * 100))
    }, headers)
  }

  async function uploadMultipartFile(
    resourceId: string,
    init: UploadInit,
    file: File,
    onProgress?: (pct: number) => void
  ): Promise<void> {
    const partSize = init.partSize && init.partSize > 0 ? init.partSize : multipartThreshold
    const localPartCount = Math.ceil(file.size / partSize)
    const partCount = Math.max(init.partCount && init.partCount > 0 ? init.partCount : 0, localPartCount)
    const loadedByPart = new Array<number>(partCount).fill(0)
    const completedParts: CompletedPart[] = []
    let nextPart = 1

    const reportProgress = () => {
      const loaded = loadedByPart.reduce((sum, value) => sum + value, 0)
      onProgress?.(Math.min(99, Math.round((loaded / file.size) * 100)))
    }

    const uploadOnePart = async (partNumber: number) => {
      const start = (partNumber - 1) * partSize
      const end = Math.min(start + partSize, file.size)
      const chunk = file.slice(start, end)
      const part = await call<MultipartPartURL>(
        `/api/v1/resources/${resourceId}/assets/multipart/part-url`,
        { method: 'POST', body: { uploadToken: init.uploadToken, partNumber } }
      )
      const xhr = await putBlob(part.uploadUrl, chunk, (loaded) => {
        loadedByPart[partNumber - 1] = loaded
        reportProgress()
      }, part.uploadHeaders)
      const etag = xhr.getResponseHeader('ETag')
      if (!etag) throw new Error('分片上传完成但没有返回 ETag')
      loadedByPart[partNumber - 1] = chunk.size
      completedParts.push({ partNumber, etag })
      reportProgress()
    }

    const worker = async () => {
      while (nextPart <= partCount) {
        const partNumber = nextPart++
        await uploadOnePart(partNumber)
      }
    }

    try {
      await Promise.all(Array.from({ length: Math.min(multipartConcurrency, partCount) }, () => worker()))
      completedParts.sort((a, b) => a.partNumber - b.partNumber)
      await call(
        `/api/v1/resources/${resourceId}/assets/multipart/complete`,
        { method: 'POST', body: { uploadToken: init.uploadToken, parts: completedParts } }
      )
      onProgress?.(100)
    }
    catch (error) {
      await call(`/api/v1/resources/${resourceId}/assets/multipart/abort`, {
        method: 'POST',
        body: { uploadToken: init.uploadToken },
      }).catch(() => {})
      throw error
    }
  }

  // uploadFile: init → PUT → finalize for one downloadable file of a resource.
  async function uploadFile(
    resourceId: string,
    file: File,
    label: string,
    onProgress?: (pct: number) => void
  ): Promise<ResourceAssetView> {
    const multipart = file.size > multipartThreshold
    const init = await call<UploadInit>(
      `/api/v1/resources/${resourceId}/assets`,
      { method: 'POST', body: { filename: file.name, size: file.size, multipart } }
    )
    if (init.method === 'MULTIPART') {
      await uploadMultipartFile(resourceId, init, file, onProgress)
    }
    else {
      await putWithProgress(init.uploadUrl, file, onProgress, init.uploadHeaders)
    }
    const res = await call<{ asset: ResourceAssetView }>(
      `/api/v1/resources/${resourceId}/assets/finalize`,
      { method: 'POST', body: { uploadToken: init.uploadToken, label } }
    )
    return res.asset
  }

  // uploadCover: init → PUT → finalize for the resource's (public) cover image.
  async function uploadCover(
    resourceId: string,
    file: File,
    onProgress?: (pct: number) => void
  ): Promise<{ coverAssetId: string; coverUrl: string }> {
    const init = await call<UploadInit>(
      `/api/v1/resources/${resourceId}/cover`,
      { method: 'POST', body: { filename: file.name, mime: file.type, size: file.size } }
    )
    await putWithProgress(init.uploadUrl, file, onProgress, init.uploadHeaders)
    return await call<{ coverAssetId: string; coverUrl: string }>(
      `/api/v1/resources/${resourceId}/cover/finalize`,
      { method: 'POST', body: { uploadToken: init.uploadToken } }
    )
  }

  return { uploadFile, uploadCover }
}
