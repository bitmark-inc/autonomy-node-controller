# Changelog

## [Unreleased]

### Added 

### Changed

### Removed

### Fixed

## [0.0.6] - 2021-07-05

### Fixed

- Fix the parameter name used for WebSocket command send_messages.

## [0.0.5] - 2021-06-18

### Fixed

- Set the message buffer size to 1000 to prevent blocking.

## [0.0.4] - 2021-06-15

### Fixed

- Prevent unnecessary restarts if the WebSocket connection is closed or the API server is temporarily unavailable.

## [0.0.3] - 2021-06-08

### Changed

- Include the request ID in every response
- Use the allow list to manage Bitcoin RPC usage

### Fixed

- Accept incomplete wallet descriptors for testnet
- Fix descriptors not loaded
- Fix the wrong command name used in the binding tool

## [0.0.2] - 2021-05-12

### Changed

- Renew JWT periodically for using Autonomy APIs.

### Fixed

- Gracefully shut down the controller when the WebSocket connection ends.

## [0.0.1]

### Added

- New commands: `bind`, `bind_ack`, `create_wallet`, `finish_psbt`, `set_member`, and `remove_member`.
