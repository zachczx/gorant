# RantKit

## Go tools

1. Templating - Templ
2. SQL - SQLx, pgx
3. Live Reload - Air
4. Authentication/Authorization - Keyclock, GoCloak
5. Cookies/Sessions - Gorilla Sessions
6. Compression - Brotli, Gzip

## JS tools

1. Hypermedia/AJAX - HTMX (+Extensions: Preload, Response Targets, SSE)
2. Formatting - Prettier
3. Bundling - ESBuild

## CSS tools

1. Framework - Tailwind
2. Components - DaisyUI

## Other tools

1. Docker
2. Makefile
3. Local mail testing - Maildev (Dev-only)
4. Terminal formatting - PTerm | Pretty Terminal Printer (Dev-only)
5. Auth/auth - Stytch (swapped out)

## TODO

1. Option for markdown formatting/processing for posts
2. File uploads
3. Need a way to disconnect live clients (remnant browser tabs causing panics)

## Notes for Choice of Auth

1. Stytch introduced about 250ms on localhost owing to service.CheckAuthentication()
2. With Keycloak it fell to 50ms on localhost and 15-20ms in prod

---

## Notes for Keycloak Setup

1. Setup a confidential client
   - OpenID Connect
   - Client Authentication > On
   - Service Accounts Role > On
2. Set an admin user and populate .env variables
3. Admin user must be assigned realm-admin role
4. Make sure this is off (default)
   - Login > Email as Username > Off
5. Choose One:
   - Authentication > Flows > Direct Grants:
     - Direct Grant - Conditional OTP > Disabled
   - Realm Settings > User Profile >
     - firstName > Required Field > Off
     - lastName > Required Field > Off
6. For Reset Password to provide a link back to app after reset,
   - Client > confidential-client:
     - Root URL > http://domain.name/
     - Home URL > http://domain.name/login
7. Optional Realm Settings:
   - Tokens > Set Access Token Lifespan > 30 days
   - Root URL originally /realms/grumplr/account/
8. TODO: Figure out why Keycloak Access Tokens expiry is max 259200 seconds
