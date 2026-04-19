# **Project Plan: Podcast Encyclopedia Pipeline**

## **Overview**

An automated pipeline that ingests a podcast RSS feed, processes episodes concurrently, and uses local LLMs to extract structured data. The goal is to build an "encyclopedia" or database of guests, companies, business models, and ideologies mentioned in the episodes.

## **Pipeline Architecture**

### **Phase 1: Ingestion & Parsing**

1. **Input:** Podcast RSS Feed URL.  
2. **Dispatch:** Use coroutines (e.g., Python asyncio) to process episodes individually and concurrently.  
3. **Concurrency Control:** Implement a semaphore or worker pool to limit concurrency. This is crucial for avoiding massive API bills or overwhelming local hardware when processing massive (10,000+ episode) backlogs.

### **Phase 2: Data Gathering & Transcription**

For each episode, extract the following:

1. **Metadata:** Episode description/show notes.  
2. **Transcript:**  
   * *Check:* Is a transcript already available in the RSS feed/metadata?  
   * *Fallback:* If no transcript exists, download the audio file and transcribe it locally (e.g., using Whisper.cpp or Faster-Whisper to avoid cloud costs).

### **Phase 3: Entity Extraction (The LLM Step)**

Pass the gathered metadata and transcript through an LLM to discern and extract structured information.

* **Target Entities:**  
  * **Humans:** Guests (Name, background, ideology).  
  * **Institutions:** Companies, organizations (Business models, customers).  
* **Chunking strategy:** Since the LLM just needs to do "a little bit of discerning in a small context window," chunk the transcript into smaller, manageable pieces before passing them to the LLM.

### **Phase 4: Storage & Output**

Generate two distinct artifacts per episode:

1. **Raw/Semi-Structured File:** The combined episode notes and full raw transcript.  
2. **Structured Output File:** The derived JSON/Database entry containing the "encyclopedia" profiles of the individuals and companies.

## **Technical Solutions & Recommendations**

### **The "Small Local Model" Problem**

You asked: *"What is the smallest model I can run on my local machine that can reliably produce structured output?"*  
Because you are doing entity extraction, you don't need a massive reasoning engine (like GPT-4). You need a model that strictly follows formatting rules (JSON Schema) and can identify nouns/concepts.  
Here are the best highly-quantized, small-parameter models (3B \- 8B) currently excelling at structured output:

1. **Qwen 2.5 (3B or 7B Instruct)**  
   * **Why:** Qwen 2.5 punches far above its weight class in instruction following and structured JSON output. The 3B version is lightning fast and can easily run on standard laptop RAM.  
2. **Llama 3.1 (8B Instruct)**  
   * **Why:** The reigning champion of the 8B class. It is incredibly reliable for structured generation. Running it heavily quantized (e.g., 4-bit GGUF format via llama.cpp or Ollama) requires only \~5-6GB of RAM.  
3. **Phi-3 Mini (3.8B)**  
   * **Why:** Microsoft's small model is optimized specifically to rival 7B+ models in logic and reasoning. It has a surprisingly capable context window and is highly efficient.

### **Forcing Structured Output**

To ensure these small models *never* break your pipeline with malformed text, do not rely on prompt engineering alone. Use a framework that forces **Grammar/Schema constraints**:

* **If using Ollama:** Use the format: "json" parameter in your API calls, or pass a strict JSON Schema if using the newest versions.  
* **If using llama.cpp (Python bindings):** Use the grammar parameter. You define a structural grammar (or JSON schema), and the inference engine physically prevents the model from generating tokens that don't fit that schema.  
* **Outlines / Instructor (Python Libraries):** These libraries wrap local models and guarantee that the output matches Pydantic models in Python. This is highly recommended for this specific project.