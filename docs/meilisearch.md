# Guide to setup local instance of Meilisearch

1. Follow [official instructions](https://www.meilisearch.com/docs/learn/self_hosted/getting_started_with_self_hosted_meilisearch) to install and setup Meilisearch locally
2. Download videos.json containing some sample data of some Safina Society Videos
3. Upload the data to your locally running instance of Meilisearch (replace MEILISEARCH_URL with the correct URL)
```
curl \
  -X POST 'MEILISEARCH_URL/indexes/videos/documents?primaryKey=id' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer aSampleMasterKey' \
  --data-binary @videos.json
```
4. Search the instance with an http request, or alternatively through a local deployment of Safina Society Search by setting the MEILISEARCH_API_KEY and MEILISEARCH_URL to the url of your local meilisearch instance in .env file.
```
curl \
  -X POST 'MEILISEARCH_URL/indexes/videos/search' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer aSampleMasterKey' \
  --data-binary '{ "q": "taqwa" }'
```
