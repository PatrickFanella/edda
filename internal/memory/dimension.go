package memory

// DefaultVectorDimension is the default number of components in an embedding
// vector. The value 768 matches the output of the nomic-embed-text model used
// by the Ollama provider. Callers that use a different model should override
// this when constructing their Embedder implementation.
const DefaultVectorDimension = 768
