import { platformAppConfig } from '@platform/ui/app-config'

// Resource uses its violet preset while shared theme primitives stay in
// @platform/ui, keeping product branding separate from platform semantics.
export default defineAppConfig(platformAppConfig('resource'))
