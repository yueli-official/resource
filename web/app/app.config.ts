import { platformAppConfig } from '@platform/ui/app-config'

// Resource site theme = the 'resource' preset (violet). Shared neutral/card/icons
// live in @platform/ui. See flightdeck/specs/2026-06-21-platform-ui-theme-layer.md.
export default defineAppConfig(platformAppConfig('resource'))
