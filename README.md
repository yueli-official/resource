# Resource product

- Lifecycle: active reusable product type
- Authority: Catalog product type `resource`, `api/` migrations/OpenAPI, `web/` UI
- Consumers: Resource site instances such as `resource-main`
- Verify: `pnpm platformctl verify product --file catalog/overlays/local.yaml --root . resource`

Resource owns downloadable resource metadata, publication, taxonomy, delivery
references and settings. Files remain Asset objects and entitlement/payment
behavior remains in Commerce. `api/` owns domain state; `web/` owns public and
management experiences.
