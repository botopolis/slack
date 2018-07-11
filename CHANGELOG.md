# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic Versioning](http://semver.org/spec/v2.0.0.html).

## [Unreleased](https://github.com/botopolis/slack/compare/v0.5.0...master)

## [0.5.0](https://github.com/botopolis/slack/compare/v0.4.0...v0.5.0)

### Added

- Reactions: `Adapter.React(bot.Message)`. React to messages with emoji.

### Changed

- Direct messages to the bot are treated as messages addressed to the bot
  (prefixed with `@{username}` and included in bot.Respond)

## [0.4.0](https://github.com/botopolis/slack/compare/v0.3.1...v0.4.0) - 2018-06-24

### Added
- Debug logging from `nlopes/slack`

### Changed
- Updated to `botopolis/bot@v0.4.0` from `v0.3.0`

### Removed

## [0.3.1](https://github.com/botopolis/slack/compare/v0.3.0...v0.3.1) - 2018-06-20

### Added

### Changed
- Fix bug causing new DMs not to be opened
- API call is made to slack on connect to get user info (vs relying on the
  websocket connection provide that information)

### Removed


## [0.3.0](https://github.com/botopolis/slack/compare/v0.2.2...v0.3.0) - 2018-06-20

### Added

- `Store.UserByEmail(email string)` for quick user lookup

### Changed

- Depends on `nlopes/slack@v0.2.0` (upgraded from a loose SHA)

### Removed
