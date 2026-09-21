<p align="center">
  <img src="assets/logo.svg" alt="LEXIO logo" width="220" />
</p>
<h1 align="center">LEXIO</h1>
<p align="center">
  A browser-native multilingual document narration engine featuring synchronized highlighting and AI-assisted execution controls.
</p>
<p align="center">
  <a href="https://voxion-labs.github.io/LEXIO">Live Environment</a>
  |
  <a href="https://github.com/liambrooks-lab/LEXIO">Repository</a>
</p>

---

## Overview
LEXIO operates as a browser-native execution environment for document narration and text processing. It integrates multi-format document ingestion, real-time text synchronization, language/accent filtering, deterministic playback controls, and AI-generated structural metadata within a single isolated workspace.

---

## System Definition
LEXIO functions as a frontend narration architecture for users requiring a centralized environment for document processing, auditory review, and content extraction.

At the technical level, LEXIO acts as:
- a multilingual document parsing engine
- a browser-native AI-assisted text-to-speech (TTS) interface
- a structured office-file narration environment
- a document analysis tool integrating summaries, glossaries, and data extraction

---

## Architectural Objective
Standard text-to-speech implementations exhibit severe deficits during actual document processing:

- limited support beyond raw plaintext
- structural degradation of imported files
- non-deterministic and robotic playback latency
- lack of visual synchronization with auditory output
- absence of integrated metadata extraction

LEXIO mitigates these deficits by providing a structured end-to-end processing pipeline:
- ingest raw content or supported file formats
- normalize data into sequentially readable blocks
- apply voice filtering via language, accent, and persona parameters
- execute playback with real-time active-word synchronization
- generate AI-structured outputs for executive briefing and content review

---

## Links
- **[View LEXIO Live](https://voxion-labs.github.io/LEXIO/)**
- **[GitHub Repository](https://github.com/liambrooks-lab/LEXIO)**
- **[Telemetry & Issues](https://github.com/voxion-labs/LEXIO/issues)**

---

## Current Project State
LEXIO currently ships with:

- an authenticated workspace initialization sequence
- distinct `Dashboard`, `Reader`, and `AI Studio` execution views
- browser speech synthesis integrating multilingual voice parameters
- live document rendering with interactive word-level playback mapping
- AI execution modes including `Executive`, `Study`, `Presentation`, and `Accessibility`
- algorithmic text normalization and natural chunking protocols
- auxiliary panels for `User Manual`, `Notes`, and `AI Assistant`
- document telemetry tracking words, sections, characters, and estimated completion time
- parsing engines for office, text, markup, and ebook format types

---

## Core Specifications
- structured workspace initialization and isolated UI
- real-time active-word synchronization during audio playback
- dynamic voice allocation based on language and accent parameters
- drag-and-drop ingestion for standard document formats
- AI-generated executive briefs, structural outlines, and review matrices
- isolated local data storage for user notes and assistant context
- responsive UI rendering for variable display environments
- modular JavaScript architecture separating auth, speech, AI, and parsing utilities

---

## Execution Surfaces
### Workspace Experience
- initialization sequence featuring profile configuration
- environment routing across `Dashboard`, `Reader`, and `AI Studio`
- global telemetry chips displaying identity, voice state, and AI confidence
- keyboard command bindings for rapid environment navigation

### Reading Experience
- source input buffer for manual text entry
- document normalization engine that renders synchronized reading sections
- playback state controls (start, pause, resume, stop)
- focus execution modes with real-time progress telemetry

### AI Assist Experience
- executive brief extraction
- structural study note generation
- presentation flow mapping
- automated glossary compilation
- entity and topic extraction
- review prompt and Q&A matrix generation

---

## Telemetry & UI State Gallery
<table>
  <tr>
    <td width="50%" valign="top">
      <img src="docs/readme/lexio-login.png" alt="LEXIO initialization experience" />
      <br />
      <strong>1. Environment Initialization</strong>
      <br />
      LEXIO initiates with a secure workspace configuration sequence, establishing user identity before loading the primary reading environment.
    </td>
    <td width="50%" valign="top">
      <img src="docs/readme/lexio-dashboard.png" alt="LEXIO dashboard view" />
      <br />
      <strong>2. Dashboard & System Overview</strong>
      <br />
      The dashboard aggregates AI-generated metadata, system commands, and live document telemetry to provide a centralized view of the current workspace state.
    </td>
  </tr>
  <tr>
    <td width="50%" valign="top">
      <img src="docs/readme/lexio-reader.png" alt="LEXIO reader view" />
      <br />
      <strong>3. Reader & Narration Engine</strong>
      <br />
      The primary execution view isolates content ingestion, document rendering, voice parameter tuning, and synchronized playback within a strict layout.
    </td>
    <td width="50%" valign="top">
      <img src="docs/readme/lexio-ai-studio.png" alt="LEXIO AI Studio view" />
      <br />
      <strong>4. AI Studio & Metadata Inspection</strong>
      <br />
      An isolated environment for inspecting generated glossaries, structural outlines, and review prompts without obstructing the primary narration pipeline.
    </td>
  </tr>
</table>

---

## Operating Directives
LEXIO is engineered to execute the following directives:
- ingest complex document structures
- normalize content for sequential processing
- optimize auditory playback latency and natural pacing
- extract and structure underlying metadata

These directives govern the UI layout, the speech synchronization loop, the parsing engine, and the AI integration layer.

---

## Linguistic & Format Specifications
LEXIO enforces multilingual narration mapping and multi-format parsing for real-world document ingestion.

### Narration Languages
- English
- Hindi
- Arabic
- Italian
- Spanish

### Supported Ingestion Formats
- PDF
- DOCX
- DOC
- PPTX
- PPT
- XLSX
- EPUB
- TXT
- Markdown
- CSV
- JSON
- RTF
- HTML
- HTM
- XML
- ODT
- ODP

---

## Tech Stack
### Core Frontend
- HTML5
- CSS3
- JavaScript (ES6+)

### Browser APIs
- Web Speech API
- DOM APIs
- Local Storage

### Parsing Engines
- PDF.js
- Mammoth.js
- JSZip
- SheetJS

### UI Resources
- Font Awesome
- Google Fonts

---

## Project Structure
```text
LEXIO/
|- assets/
|- css/
|  `- components.css
|- docs/
|  |- API.md
|  |- CHANGELOG.md
|  |- CONTRIBUTING.md
|  `- readme/
|       
|     
|- js/
|  |- ai.js
|  |- auth.js
|  |- fileHandler.js
|  |- speech.js
|  `- utils.js
|- index.html
|- LICENSE
|- package.json
|- README.md
|- script.js
`- styles.css
```

---

## Architecture
### System Responsibilities
- workspace initialization and state management
- document normalization and active telemetry
- playback execution and active-word synchronization
- auxiliary panel management (notes, manual, AI)
- dashboard and AI studio state rendering

### Ingestion and Execution Flow
1. User supplies raw text or supported file binaries.
2. LEXIO routes the file to the appropriate parsing engine.
3. Content is normalized into a sequential array of readable blocks.
4. The document viewer maps text elements to clickable DOM nodes.
5. The speech engine executes playback while triggering real-time highlighting callbacks.
6. The AI module processes the text payload to generate metadata, glossaries, and structural prompts.

### Speech Subsystem
- voice allocation is strictly filtered by designated language, accent, and persona variables
- sequential text is algorithmically chunked to optimize pacing and prevent synthesis timeouts
- playback state and active DOM nodes maintain strict synchronization
- output relies on native browser and OS-level TTS synthesizer capabilities

---

## Validation Snapshot
The current repository state includes:
- modular JavaScript architecture isolated by concern (`auth`, `speech`, `fileHandler`, `ai`, `utils`)
- executable via local static server (`live-server`)
- functional binary parsing across multiple document formats
- initial testing infrastructure placeholder via npm scripts

---

## Key Capabilities
- structured execution workspace
- multilingual document synthesis
- active-word synchronized playback
- multi-format binary ingestion (Office, EPUB, PDF)
- AI-generated structural metadata and review modes
- isolated local data storage for notes and contextual processing
- keyboard-driven command navigation
- deployable as a static client-side application

---

## Current Deployment Scope
LEXIO is currently deployed as a functional frontend architecture featuring:
- multi-format document ingestion
- multilingual TTS integration
- distinct execution environments (Dashboard, Reader, AI Studio)
- AI-driven structural metadata extraction
- localized state management for identity and note-taking
- static web deployment

---

## Local Setup
### Prerequisites
- Node.js `14+`
- npm `6+`
- A modern browser with Web Speech API support

### Install Dependencies
```bash
npm install
```

### Run Local Execution Server
```bash
npm run dev
```

### Alternate Initialization
```bash
npm start
```

### Local Endpoint
- `http://localhost:3000`

---

## Build Directives
### Start Development Environment
```bash
npm run dev
```

### Start Local Preview
```bash
npm start
```

### Execute Test Suite
```bash
npm test
```

---

## Deployment
### Live Deployment
- Hosted as a static client-side application
- Public deployment URL: [[https://voxion-labs.github.io/LEXIO/](https://voxion-labs.github.io/LEXIO/)]

### Deployment Profile
- lightweight browser-native frontend
- operates completely client-side without backend dependencies
- speech synthesis execution relies on local client capabilities

---

## License

LEXIO operates under a custom restricted proprietary license.

The full license text is available in the [LICENSE](LICENSE) directive.

License summary:

- copyright © 2026 Rudranarayan Jena
- all rights reserved
- unauthorized copying, modification, distribution, hosting, reuse, or derivative work creation is strictly prohibited
- commercial or non-commercial exploitation is forbidden unless explicitly authorized in writing by the architect

LEXIO is **not** an open-source project and is not distributed under MIT, Apache, GPL, or any other permissive license.

---

## Architecture & Infrastructure
<p align="center">
  <img src="docs/readme/author-rudranarayan-jena.jpg" alt="Rudranarayan Jena" width="180" />
</p>
<p align="center">
  <strong>Maintained by <a href="https://github.com/liambrooks-lab">Rudranarayan Jena</a></strong>
</p>
<p align="center">
  <strong>Founder of Voxion Labs</strong>

---
