<script setup lang="ts">
import { ManageAssetSettings } from '@platform/asset/components'

definePageMeta({ layout: 'manage', middleware: 'auth' })

const { slug, brand } = useSiteRuntime()
const { can } = useResourceMe()
const canManageAssets = computed(() => can('resource.asset_settings.manage'))
useSeoMeta({ title: '资源配置 · 控制台' })
</script>

<template>
  <ManageAssetSettings
    :site-key="slug"
    :site-name="brand"
    :can-manage="canManageAssets"
    permission-description="资源配置会影响全站上传、存储与访问级别，请使用站点管理员账户操作。"
    description="配置资源站使用资源中心时的默认存储、封面、正文图片与下载文件规则。"
    :profile-order="['resource-cover', 'resource-content', 'resource']"
    delivery-profile-key="resource"
  />
</template>
