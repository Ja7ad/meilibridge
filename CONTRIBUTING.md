# Contributing to Meilibridge

Thank you for considering contributing to Meilibridge! We welcome contributions from the community to help improve and 
enhance this project. Whether you're fixing bugs, adding new features, or improving documentation, your 
help is greatly appreciated.

## Getting Started

### Prerequisites
- Make sure you have [Go](https://golang.org/doc/install) installed.
- Familiarize yourself with [Meilisearch](https://www.meilisearch.com) and 
the [Meilibridge documentation](https://github.com/Ja7ad/meilibridge).

### Fork the Repository
1. Fork the Meilibridge repository by clicking the "Fork" button on the GitHub page.
2. Clone your forked repository:
   ```sh
   git clone https://github.com/your-username/meilibridge.git
   cd meilibridge
   ```

### Set Up Your Development Environment
1. Install the dependencies:
   ```sh
   go mod tidy
   ```

2. Run the tests to ensure everything is working:
   ```sh
   go test ./...
   ```

## Making Changes

### Branching
1. Create a new branch for your feature or bug fix:
   ```sh
   git checkout -b your-feature-branch
   ```

### Coding Style
- Follow the [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments) and [Effective Go](https://golang.org/doc/effective_go.html) guidelines.
- Ensure your code is properly formatted using `gofmt`:
  ```sh
  gofmt -w .
  ```

### Commit Messages
- Write clear and concise commit messages.
- Use the present tense ("Add feature" not "Added feature").
- Include the issue number if applicable (e.g., `Fix #123: Add new feature`).

### Testing
- Add tests for any new functionality you add.
- Ensure all tests pass before submitting your changes:
  ```sh
  go test ./...
  ```

## Submitting Changes

1. Push your branch to your forked repository:
   ```sh
   git push origin your-feature-branch
   ```

2. Open a pull request (PR) against the `main` branch of the [Meilibridge repository](https://github.com/Ja7ad/meilibridge):
   - Provide a clear title and description of your changes.
   - Reference any related issues or pull requests.
   - Include any relevant screenshots or logs if your changes include UI or significant alterations.

### Code Review
- Your pull request will be reviewed by other contributors.
- Be prepared to make changes based on feedback.
- Once your pull request is approved, it will be merged into the main branch.

## Documentation
- If your contribution adds or changes functionality, update the documentation accordingly.
- Ensure the `README.md` and any other relevant documents are up-to-date.

## Reporting Issues
- If you encounter any issues, please [open an issue](https://github.com/Ja7ad/meilibridge/issues) on GitHub.
- Provide as much detail as possible, including steps to reproduce the issue, your environment, and 
- any relevant logs or screenshots.

## Thank You!
Thank you for contributing to Meilibridge! Your efforts help make this project better for everyone. We look 
forward to working with you.

For more information, feel free to contact the maintainers or refer to the existing documentation.

Happy coding!