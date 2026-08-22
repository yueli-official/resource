import { optimizeImageFile } from '@yueli/ui/image/browser'
import { assetUploadURL } from '@yueli/asset-nuxt/upload'
import type { AssetView } from '~/types'
import { publicAssetMediaUrl } from '~/utils/asset-media.mjs'

type UploadInit = {
  uploadUrl: string
  uploadToken: string
  method?: 'PUT' | 'MULTIPART' | string
  uploadHeaders?: Record<string, string>
  uploadId?: string
  partSize?: number
  partCount?: number
}

type MultipartPartURL = {
  uploadUrl: string
  uploadHeaders?: Record<string, string>
}

type CompletedPart = {
  partNumber: number
  etag: string
}

interface UploadOptions {
  profileKey?: string
  visibility?: 'public' | 'private'
  deliveryPolicy?: string
  onProgress?: (pct: number) => void
  optimizeImage?: boolean
  imageMaxSide?: number
  imageQuality?: number
}

export function useAssetUpload() {
  const { call } = useAssetApi()
  const config = useRuntimeConfig()
  const multipartThreshold = 64 * 1024 * 1024
  const multipartConcurrency = 3

  function putBlob(url: string, body: Blob, onLoaded?: (loaded: number, total: number) => void, headers?: Record<string, string>): Promise<XMLHttpRequest> {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest()
      xhr.open('PUT', assetUploadURL(url))
      for (const [key, value] of Object.entries(headers ?? {})) xhr.setRequestHeader(key, value)
      xhr.upload.onprogress = (event) => {
        if (event.lengthComputable) onLoaded?.(event.loaded, event.total)
      }
      xhr.onload = () => {
        if (xhr.status >= 200 && xhr.status < 300) resolve(xhr)
        else reject(new Error(`上传失败 (HTTP ${xhr.status})`))
      }
      xhr.onerror = () => reject(new Error('上传网络错误，请确认资源服务在线'))
      xhr.send(body)
    })
  }

  function putWithProgress(url: string, file: File, onProgress?: (pct: number) => void, headers?: Record<string, string>) {
    return putBlob(url, file, (loaded, total) => {
      onProgress?.(Math.max(1, Math.round((loaded / total) * 100)))
    }, headers)
  }

  async function uploadMultipartFile(init: UploadInit, file: File, onProgress?: (pct: number) => void) {
    const partSize = init.partSize && init.partSize > 0 ? init.partSize : multipartThreshold
    const localPartCount = Math.ceil(file.size / partSize)
    const partCount = Math.max(init.partCount && init.partCount > 0 ? init.partCount : 0, localPartCount)
    const loadedByPart = new Array<number>(partCount).fill(0)
    const completedParts: CompletedPart[] = []
    let nextPart = 1

    const reportProgress = () => {
      const loaded = loadedByPart.reduce((sum, value) => sum + value, 0)
      onProgress?.(Math.min(99, Math.max(1, Math.round((loaded / file.size) * 100))))
    }

    const uploadOnePart = async (partNumber: number) => {
      const start = (partNumber - 1) * partSize
      const end = Math.min(start + partSize, file.size)
      const chunk = file.slice(start, end)
      const part = await call<MultipartPartURL>('/api/v1/assets/multipart/part-url', {
        method: 'POST',
        body: { uploadToken: init.uploadToken, partNumber }
      })
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
        const partNumber = nextPart
        nextPart += 1
        await uploadOnePart(partNumber)
      }
    }

    try {
      await Promise.all(Array.from({ length: Math.min(multipartConcurrency, partCount) }, () => worker()))
      completedParts.sort((a, b) => a.partNumber - b.partNumber)
      await call('/api/v1/assets/multipart/complete', {
        method: 'POST',
        body: { uploadToken: init.uploadToken, parts: completedParts }
      })
      onProgress?.(100)
    } catch (error) {
      await call('/api/v1/assets/multipart/abort', {
        method: 'POST',
        body: { uploadToken: init.uploadToken }
      }).catch(() => {})
      throw error
    }
  }

  function normalizeUploadError(err: unknown) {
    const message = (err as Error).message || ''
    if (message.includes('asset.upload.too_large') || message.includes('file too large')) {
      return new Error('素材超过当前用途的大小上限，请压缩后重试。')
    }
    if (message.includes('asset.upload.invalid_type') || message.includes('file type not allowed')) {
      return new Error('这个文件类型不在当前用途允许范围内。')
    }
    return err instanceof Error ? err : new Error(message || '上传失败')
  }

  async function initUpload(file: File, options: UploadOptions, multipart: boolean) {
    return await call<UploadInit>('/api/v1/assets/upload-init', {
      method: 'POST',
      body: {
        filename: file.name,
        mime: file.type || 'application/octet-stream',
        size: file.size,
        siteKey: config.public.assetNamespace,
        profileKey: options.profileKey || 'resource-content',
        category: options.profileKey || 'resource-content',
        visibility: options.visibility || 'public',
        deliveryPolicy: options.deliveryPolicy || '',
        multipart
      }
    })
  }

  async function uploadAsset(rawFile: File, options: UploadOptions = {}): Promise<AssetView> {
    try {
      const file = await optimizeImageFile(rawFile, {
        enabled: options.optimizeImage === true,
        maxSide: options.imageMaxSide,
        quality: options.imageQuality,
        outputType: 'image/webp'
      })
      const wantsMultipart = file.size > multipartThreshold
      let init: UploadInit
      try {
        init = await initUpload(file, options, wantsMultipart)
      } catch (err) {
        const message = (err as Error).message || ''
        if (!wantsMultipart || (!message.includes('multipart unsupported') && !message.includes('storage: multipart unsupported'))) {
          throw err
        }
        init = await initUpload(file, options, false)
      }

      if (init.method === 'MULTIPART') await uploadMultipartFile(init, file, options.onProgress)
      else await putWithProgress(init.uploadUrl, file, options.onProgress, init.uploadHeaders)

      options.onProgress?.(99)
      const finalized = await call<{ asset: AssetView }>('/api/v1/assets/finalize', {
        method: 'POST',
        body: { uploadToken: init.uploadToken }
      })
      options.onProgress?.(100)
      return finalized.asset
    } catch (err) {
      throw normalizeUploadError(err)
    }
  }

  async function uploadPublicImage(file: File, onProgress?: (pct: number) => void, profileKey = 'resource-content') {
    const asset = await uploadAsset(file, {
      profileKey,
      visibility: 'public',
      onProgress,
      optimizeImage: true,
      imageMaxSide: profileKey === 'resource-cover' ? 1600 : 1920,
      imageQuality: profileKey === 'resource-cover' ? 0.9 : 0.86
    })
    return {
      asset,
      url: publicAssetMediaUrl(asset.id, profileKey === 'resource-cover' ? 'detail' : 'content')
    }
  }

  return { uploadAsset, uploadPublicImage }
}
