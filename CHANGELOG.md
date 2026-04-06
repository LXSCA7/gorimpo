# 📜 Changelog

All notable changes to this project will be documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]
### Added
- Initial planning for `gorimpo-docs` website.
- Preparation for OLX Native API integration.
- Full Localization: Standardized all system logs, error and internal notification strings, and Prometheus metrics to English (by @dipievil & @qorexdev).

---

## [v1.2.1] - 2026-04-05
### 🛠️ Summary
This version focuses on **resilience and modularity**. We introduced dynamic proxy support to avoid IP bans, refactored the OLX adapter for a cleaner architecture, and implemented Gotify notifications.

### 🚀 Added
- **Proxy Rotation:** Dynamic rotation with initial ProxyScrape support.
- **Gotify Support:** Native integration for Gotify notifications (by @dipievil).
- **Hexagonal Refactor:** Moved OLX scraper to a modular sub-package in `/internal/adapters/scraper/olx/`.

### 🔧 Fixed & Improved
- **Docker Performance:** Optimized Dockerfile using multi-stage builds (by @dipievil).
- **Circuit Breaker:** The system now ignores proxy timeouts to prevent unnecessary cooldowns.
- **I18n:** Initial structure for internationalization (EN/PT-BR).