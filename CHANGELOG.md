# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](http://keepachangelog.com/en/1.0.0/)
and this project adheres to [Semantic
Versioning](http://semver.org/spec/v2.0.0.html).

## Unreleased

## [1.6.0] - 2020-11-05

### Changed
- Added --include-labels boolean as option to include entity and check labels, false by default
- Removed sensu_entity_name and sensu_check_name labels

## [1.5.0] - 2020-11-05

### Changed
- Updated Sensu SDK to 0.11.0
- Reorganized options handling to meet current standard
- Replace all "/", ".", and "-" in label names with "_"

## [1.3.0] - 2020-02-18
### Changed
- Another release to test goreleaser

## [1.2.0] - 2020-02-17
### Added
- Adding Sensu entity and check labels to Stackdriver metric series labels

## [1.1.0] - 2020-02-17
### Added
- Including a Sensu check name label on Stackdriver metrics

## [1.0.1] - 2020-02-11
### Changed
- README edits for Bonsai

## [1.0.0] - 2020-02-11
### Added
- Metric time series chunk support, making a Stackdriver request for every
200 time series.

## [0.0.3] - 2020-02-11
### Changed
- Dropping event metric points after 200 time series (the stackdriver request maximum)

## [0.0.2] - 2020-02-10

### Fixed
- Removed OS X i386 from Bonsai configuration

## [0.0.1] - 2020-02-10

### Added
- Initial release
