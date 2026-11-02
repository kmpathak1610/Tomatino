# Implementation Status

| Module | Status | Implemented | Tested | Remaining |
|---|---|---|---|---|
| Monorepo scaffolding | IMPLEMENTED | yes | no | - |
| Backend core (Go) | IMPLEMENTED | yes | partial | more modules |
| Database schema/migrations | TESTED | yes | yes | PostGIS container verified |
| Auth (OTP dev) | IMPLEMENTED | yes | no | JWT refresh, CSRF |
| Restaurants | IMPLEMENTED | yes | no | onboarding docs |
| Menu | IMPLEMENTED | yes | no | variants/addons |
| Cart | IMPLEMENTED | yes | no | coupon application |
| Coupons | IMPLEMENTED | yes | no | redemption tracking |
| Orders | IMPLEMENTED | yes | yes (transitions) | concurrency tests |
| Payments | IMPLEMENTED (mock) | yes | no | Razorpay sandbox |
| Delivery | IMPLEMENTED | yes | no | location retention |
| Reviews | IMPLEMENTED | yes | no | delivery reviews |
| Support | IMPLEMENTED | yes | no | admin ticket mgmt |
| Notifications | IMPLEMENTED | yes | no | push/SMS adapters |
| Favourites | IMPLEMENTED | yes | no | - |
| Device Tokens | IMPLEMENTED | yes | no | - |
| Partner APIs | IMPLEMENTED | yes | no | staff/branch mgmt |
| Admin APIs | IMPLEMENTED | yes | no | settlements/promotions |
| Customer web | IMPLEMENTED | yes | no | polish, filters |
| Admin web | IMPLEMENTED | yes | no | finance/promotions UI |
| Restaurant web | IMPLEMENTED | yes | no | add-item form, settings |
| Delivery web | IMPLEMENTED | yes | no | maps, location |
| Docker compose | IMPLEMENTED | yes | yes | postgis verified |
| Tests (backend) | IN_PROGRESS | unit | yes | integration/e2e |
| Realtime (WebSockets) | NOT_STARTED | no | no | full module |
| Finance/settlements | NOT_STARTED | no | no | full module |
| E2E Playwright | NOT_STARTED | no | no | full suite |
| Docs | IMPLEMENTED | yes | n/a | expand |
