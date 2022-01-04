# Rebuilding LNDHub

Goal of this project is to build a simple accounting system with a [LNDHub](https://github.com/BlueWallet/LndHub) compatible API that focusses on simplicity, maintainability and ease of deployment.

[LNDHub](https://github.com/BlueWallet/LndHub) is a simple accounting system for LND. It allows users to send and receive lightning payments. Through the API people can access funds through a shared lightning node. (see overview.png diagram)

Some design goals:

* No runtime dependencies (all compiled into a single, simple deployable executable)
* Use of an ORM ([gorm.io](https://gorm.io/)?)to support deployments with SQLite and PostgreSQL (default) as databases
* Focus on offchain payments (no onchain transactions supported)
* Plan for multiple node backends ([LND](https://github.com/lightningnetwork/lnd/) gRPC interface is the first implementation) (also through Tor)
* Admin panel for better Ops
* All configuration stored in the DB



### API endpoints

See [LNDHub API](https://github.com/BlueWallet/LndHub/blob/master/controllers/api.js) for enpoints and request/response signatures.

#### /create
Create a new user account

#### /auth
Get new "session" access/refresh tokens. access token is required for all other API endpoints

#### /addinvoice
Generate a new lightning invoice

#### /payinvoice
Pay a lightning invoice

#### /checkpayment/:payment_hash
Check the status of an incoming transaction

#### /balance
Get the user's balanc

#### /gettxs
Get all transactions



### ToDos

- [ ] Project setup for [Echo](https://echo.labstack.com/), [gorm](https://gorm.io/) (with support for PostgreSQL and SQLite), Unit-Test setup
- [ ] Implement first endpoints (`/create`, `/auth`, `/addinvoice`)
- [ ] Connect to LND (gRPC API) (in the future the API implementation should be configurable)
- [ ] ...


### Datamodel

* Double entry accounting?
	+ https://gist.github.com/NYKevin/9433376
	+ https://gocardless.com/guides/posts/double-entry-bookkeeping/
	+


### Links

* [LNDHub](https://github.com/BlueWallet/LndHub) - Current nodejs implementation