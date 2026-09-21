# LEXIO Operational & Contribution Protocols

LEXIO operates under strict execution and rendering protocols. We do not accept arbitrary UI bloat, unoptimized DOM manipulation, or synchronous blocking logic. This repository is maintained for high-performance, browser-native document narration.

If you intend to submit a Pull Request, you must adhere strictly to the following institutional directives.

## 1. Architectural Standards
All code submitted to LEXIO must meet our baseline performance metrics:
* **Zero-Blocking Event Loop:** Submissions causing main-thread jank during document ingestion or TTS playback will be instantly rejected. Large document parsing (PDF, EPUB) must utilize Web Workers or strict asynchronous chunking.
* **Deterministic DOM Memory:** Uncontrolled event listener attachments, detached DOM nodes, and memory leaks during large document rendering are strictly prohibited. Prove your garbage collection (GC) stability via telemetry or memory profiles before submission.
* **Speech API Efficiency:** Do not rely on bloated external TTS libraries. We interface directly with native Browser/OS capabilities. Narration chunking algorithms must maintain sub-15ms buffer transitions to prevent audio stutter.

## 2. Pull Request (PR) Governance
Before initiating a merge request, ensure your PR adheres to this exact structure:
1. **[METRIC] Benchmark Data:** You must provide before/after execution telemetry (e.g., document parse time, TTS audio buffer latency, DOM node count).
2. **[LOGIC] State Transition:** Explicitly document the architectural state transitions your code alters (e.g., playback synchronization, AI context windows).
3. **[ISOLATION] Threat Model:** For AI and document parsing modules, prove that no Cross-Site Scripting (XSS) vulnerabilities or malicious payload executions are possible via manipulated source files.

*Note: PRs failing to provide empirical benchmark data will be closed immediately without review.*

## 3. Vulnerability Disclosure
**DO NOT** open public issues for zero-day exploits, document parsing XSS vulnerabilities, or prompt-injection bypasses within the AI Studio. Public disclosure of critical threats compromises the integrity of the engine.
* All security reports must be routed internally.
* Contact the Lead Architect directly for secure transmission protocols.

## 4. Code of Conduct
We evaluate code, not intentions. Your submissions will be scrutinized ruthlessly based on algorithmic efficiency and browser rendering optimization. Keep discussions clinical, objective, and exclusively focused on frontend system architecture.
