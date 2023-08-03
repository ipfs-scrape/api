# IPFS Scrape API

This is a API that returns the scraped IPFS data from a DynamoDB table. The api is implemented in Go and uses the Gin web framework.

## Configuration

The api runs on port `8080`
The api is configured using the following environment variables:

- `IPFS_DYNAMODB_NAME`: The name of the DynamoDB table to use.

## Endpoints

The api exposes the following endpoints:

- `GET /ping`: Returns `"pong"`.
- `GET /tokens`: Returns all items in the DynamoDB table.
- `GET /tokens/:cid`: Returns the item with the specified CID.
- `POST /tokens/:cid`: Adds the specified CID to the queue for processing.
- `POST /bulk`: Adds the specified CIDs to the queue for processing. Expects a JSON payload with a `cids` array.
- `POST /csv`: Adds the CIDs in the specified CSV file to the queue for processing. Expects a `multipart/form-data` payload with a `file` field containing the CSV file.

## Dependencies

The api uses the following dependencies:

- `github.com/aws/aws-sdk-go/aws/session`
- `github.com/aws/aws-sdk-go/service/dynamodb`
- `github.com/gin-gonic/gin`
- `github.com/gokitcloud/ginkit`
- `github.com/ipfs-scrape/api/backend`
- `github.com/ipfs-scrape/api/ipfs`
- `github.com/ipfs-scrape/api/queue`
- `github.com/sirupsen/logrus`

## License

This code is licensed under the MIT License. See the `LICENSE` file for details.
