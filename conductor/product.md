# Initial Concept
An automated pipeline that ingests a podcast RSS feed, processes episodes concurrently, and uses local LLMs to extract structured data. The goal is to build an "encyclopedia" or database of guests, companies, business models, and ideologies mentioned in the episodes.

# Product Guide

## Overview
Podpedia is a podcast analysis tool designed to build a structured encyclopedia of podcast content. By processing RSS feeds and using local LLM inference, it identifies and extracts connections between guests, companies, and business models.

## Target Audience
- **Data Researchers:** To perform market analysis or research on podcast-based insights.
- **Individual Users:** To manage and explore their own private podcast libraries.

## Core Features
- **RSS Ingestion:** Automated fetching and parsing of podcast feeds.
- **Local LLM Extraction:** Use of Ollama and OpenAI-compatible inference engines for data extraction.
- **Search & Discovery:** Tools to find guests and companies across massive episode backlogs.
- **Data Export:** Exporting findings to various formats for external use.
- **Knowledge Graph Visualization:** Visualizing the connections between people, companies, and ideas.

## Technical Goals
- **High Concurrency:** Efficiently processing 10,000+ episodes without excessive resource use.
- **Local-First:** Prioritizing local hardware for privacy and cost control.
- **Markdown-Based Storage:** Storing the final encyclopedia as a collection of structured Markdown files for portability and ease of search.
