# daily-planet

Self-hosted Discord bot for managing RSS feeds in a server and in DMs

## Deployment

Deployed using custom hosted Dokploy server and Docker image built from Dockerfile.

The image is automatically built and deployed using `.github/workflows/publish_docker_image.yaml` to [GitHub Container Registry](https://docs.github.com/en/packages/working-with-a-github-packages-registry/working-with-the-container-registry)
To setup the project on Dokploy:

1. Create new project
2. Add a new "Application" service
3. Set the provider to "Docker"
  - Docker image: `ghcr.io/woojiahao/daily_planet:main`
  - Registry URL: `ghcr.io`
4. Add `DISCORD_TOKEN` and `DB_ROOT` to "Environment"
  - `DB_ROOT` is recommended to be `/data/`
5. Create a volume mount under "Advanced"
  - Container path must be the same as `DB_ROOT`

## Running on Docker

```bash
git clone https://github.com/woojiahao/daily-planet.git
cd daily-planet
docker build --no-cache -t daily_planet .
docker run \
  --rm \
  --env DISCORD_TOKEN="<token>" \
  --env DB_ROOT=/data/ \
  -v "<volume path>":/data \
  -i \
  -t \
  daily_planet:latest
```
