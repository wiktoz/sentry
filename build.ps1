# build.ps1

# Variables
$goBuilderImage = "builder-go-native"
$webBuilderImage = "builder-web-native"
$finalImage = "wiktoz/sentry:armv6"
$finalDockerfile = "final.Dockerfile"

# Clean up any old temp containers
docker container rm -f extract-go
docker container rm -f extract-web

Write-Host "Step 1: Build Go builder stage (native)..."
docker build --target builder-go -t $goBuilderImage .

Write-Host "Step 1: Build Web builder stage (native)..."
docker build --target builder-web -t $webBuilderImage .

# Remove previous Go binary if exists
if (Test-Path ./server) {
    Write-Host "Removing old server binary..."
    Remove-Item ./server -Force
}

# Extract Go binary
Write-Host "Extracting Go binary from builder image..."
docker create --name extract-go $goBuilderImage | Out-Null
docker cp extract-go:/app/api/server ./server
docker rm extract-go | Out-Null

# Remove previous dist folder if exists
if (Test-Path ./dist) {
    Write-Host "Removing old dist folder..."
    Remove-Item ./dist -Recurse -Force
}

# Extract Web frontend dist folder
Write-Host "Extracting Web frontend dist from builder image..."
docker create --name extract-web $webBuilderImage | Out-Null
docker cp extract-web:/app/web/dist ./dist
docker rm extract-web | Out-Null

# Check if final Dockerfile exists
if (-Not (Test-Path $finalDockerfile)) {
    Write-Host "Error: Final Dockerfile '$finalDockerfile' not found. Please create it first." -ForegroundColor Red
    exit 1
}

Write-Host "Step 2: Build final ARMv7 image..."
docker buildx build --platform linux/arm/v6 --load -f $finalDockerfile -t $finalImage .

Write-Host "Build finished. Final image tagged as '$finalImage'"
