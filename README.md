# GoRant: Go Webapp Starter Kit

My personal Go webapp starter kit.

## Go tools

1. Templating - Templ
2. SQL - SQLx, pgx
3. HMR - Air
4. Auth/auth - Stytch Go SDK
5. Auth/auth - Keyclock, GoCloak
6. Sessions - Gorilla Sessions

## JS tools

1. Hypermedia/AJAX - HTMX
2. Formatting - Prettier
3. Bundling - ESBuild

## CSS tools

1. Framework - Tailwind
2. Components - DaisyUI

## Other tools

1. Docker
2. Makefile

## To Do

1. Forgot/reset password
2. Real time posts with SSE

## Notes for Choice of Auth

1. Stytch introduced about 250ms on localhost owing to service.CheckAuthentication()
2. With Keycloak it fell to 50ms on localhost and 15-20ms in prod

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
   - Disable all required actions in Authentication
   - Realm Settings > User Profile >
     - firstName > Required Field > Off
     - lastName > Required Field > Off
6. Optional Realm Settings:
   - Tokens > Set Access Token Lifespan > 30 days
7. TODO: Figure out why Keycloak Access Tokens expiry is max 259200 seconds
