package pipeline_test

// GoTimeTranscriptFixture is a representative transcript for testing extraction.
// Taken from a sample of Go Time #320.
const GoTimeTranscriptFixture = `
[00:00:00] Mat Ryer: Hello, and welcome to Go Time. I'm Mat Ryer. Today we're talking about systems programming in Go. Joining me today is Kris Brandow. Hello, Kris.
[00:00:12] Kris Brandow: Hey, Mat. Great to be here.
[00:00:15] Mat Ryer: And our special guest today is Alice Smith from Acme Corp. Alice is a lead systems engineer who's been building a new distributed database in Go. Welcome, Alice.
[00:00:25] Alice Smith: Thanks, Mat. Yeah, we've been working on this project called 'GopherDB' at Acme for about two years now.
[00:00:32] Mat Ryer: GopherDB? That sounds interesting. What was the motivation behind building a new database in Go instead of using something like C++ or Rust?
[00:00:40] Alice Smith: Well, at Acme Corp, we really value developer productivity and safety. C++ gives you the performance, but the safety guarantees aren't there. Rust is great, but the learning curve for our team was a bit steep. Go provided that sweet spot of high performance with excellent concurrency primitives like goroutines and channels, which are perfect for a database engine.
[00:01:02] Kris Brandow: And Alice, I saw your talk at GopherCon where you mentioned that GopherDB is primarily targeting SMBs (Small to Medium Businesses) who need a simplified SaaS offering for their data storage.
[00:01:15] Alice Smith: Exactly. Most enterprise databases are too complex for smaller shops. We wanted to build something that 'just works'.
[00:01:22] Mat Ryer: That's fascinating. Let's dive deeper into the architecture of GopherDB...
`
