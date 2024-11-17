# portfolio-api

`portfolio-api` is a retrieval augmented generation API to allow users to ask questions about the content in a developer portfolio.

Intended to be integrated into a portfolio website.

## how it works
1. Opens `experience.json` and `projects.json` files to retrieve experiences and projects.
2. Generates vector embeddings for each experience and project, stores them in a SQLite database.
- Duplicates are ignored by checking the content hash
3. User sends `POST /api/v1/ask` request with a question
4. Calculates cosine similarity between the question and each embedding in the database
- This is currently being done in the application layer, but should be done in the database layer if the db has a large amount of embeddings
5. Top 3 most similar documents are used to generate a prompt for the LLM

## installation

### prerequisites
- [Go](https://go.dev/doc/install)
- [Docker](https://docs.docker.com/get-docker/)

### local development

1. Clone the repository
2. Initialize .env file with the respective variables
3. Run `source <(make exportenv)`
4. Run `make run/docker`