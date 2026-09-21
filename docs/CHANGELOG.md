# Changelog

All notable changes to LEXIO will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2024-03-19

###  Initial Release

#### Added
- **Authentication System**
  - User login with name, email, phone
  - Form validation
  - User avatar with dynamic colors
  - Activity logging

- **Text Input & Processing**
  - Multi-line text input
  - Real-time word/character counter
  - Reading time estimator
  - Clear text functionality

- **File Upload Support**
  - TXT file upload with drag & drop
  - PDF text extraction using PDF.js
  - File size validation (10MB limit)
  - Visual file info display

- **Speech Controls**
  - Play/Pause/Resume/Stop functionality
  - Speed control (0.5x to 3x)
  - Pitch control (premium feature)
  - Multiple voice selection (Male/Female/Child)

- **Visual Features**
  - Real-time word highlighting during playback
  - Click any word to start reading from there
  - Progress bar tracking
  - Previous/Next word navigation

- **UI/UX**
  - Three-column dashboard layout
  - Glass morphism design
  - Dark theme optimized
  - Fully responsive for all devices
  - Loading screen with animation
  - Status messages with different types

- **Enterprise Features**
  - Activity log tracking
  - Trust badges display
  - Professional footer with links
  - Stats dashboard
  - Premium feature badges

- **Documentation**
  - Complete README
  - Contribution guidelines
  - Changelog
  - API documentation template
  - PR template
  - Funding configuration

- **GitHub Integration**
  - GitHub Actions workflow for auto-deploy
  - Pages deployment ready
  - MIT License
  - .gitignore configuration

### Technical
- Modular JavaScript architecture
- CSS variables for theming
- Responsive grid layout
- Web Speech API integration
- PDF.js for PDF processing
- Zero external dependencies (except PDF.js)
- Mobile-first approach
- Cross-browser compatible

### File Structure
- 17 core files
- Organized module system
- Separate concerns (auth, speech, file handling, utils)
- Assets folder for static resources
- Documentation folder
- GitHub workflows

### Known Issues
- PDF text extraction may vary based on PDF formatting
- Voice selection limited to browser's available voices
- Export functionality (MP3/WAV) coming in v1.1
- DOCX/EPUB support coming in v1.2

### Dependencies
- PDF.js v3.11.174
- Google Fonts (Inter)

## [0.9.0] - 2024-03-01 (Beta)

### Added
- Basic text-to-speech functionality
- Simple login prototype
- Initial UI design
- Core architecture

---

For more information, visit [LEXIO on GitHub](https://github.com/Voxion-Labs/LEXIO)