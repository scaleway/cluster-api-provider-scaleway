#!/bin/bash
set -e

# Default values
BUCKET_NAME=""
TARGET_ZONES=""
TARGET_API="block"
IMAGE_NAME=""
SNAPSHOT_ID=""
SNAPSHOT_ZONE="fr-par-1"
SNAPSHOT_API="block"
PROJECT_ID=""
ARCH="x86_64"

print_help() {
  echo "Usage: $0 [options]"
  echo "Options:"
  echo " Transient resources parameters:"
  echo "  -b, --bucket-name    BUCKET_NAME     Specify a bucket name (the bucket must be in the same region as the snapshot)"
  echo " Source snapshot parameters:"
  echo "  -s, --snapshot-id    SNAPSHOT_ID     Specify the ID of the snapshot that will be used to create images"
  echo "  -z, --snapshot-zone  SNAPSHOT_ZONE   Specify the zone of the snapshot that will be used to create images"
  echo "  -a, --snapshot-api   SNAPSHOT_API    Specify the API where the snapshot is present (block or instance)"
  echo " Output images parameters:"
  echo "  -c, --arch           ARCH            Specify the CPU architecture"
  echo "  -p, --project-id     PROJECT_ID      Specify the project ID where the images will be created"
  echo "  -i, --image-name     IMAGE_NAME      Specify the name of the images to create"
  echo "  -o, --target-api     TARGET_API      Specify the target API, where snapshots will be created (block or instance)"
  echo "  -t, --target-zones   TARGET_ZONES    Specify the target zones, separated by comma"
  echo " Misc:"
  echo "  -h, --help                           Show this help message"
}

# Parse flags
while [[ "$#" -gt 0 ]]; do
  case "$1" in
    -b|--bucket-name)
      BUCKET_NAME="$2"
      shift 2
      ;;
    -t|--target-zones)
      TARGET_ZONES="$2"
      shift 2
      ;;
    -o|--target-api)
      TARGET_API="$2"
      shift 2
      ;;
    -i|--image-name)
      IMAGE_NAME="$2"
      shift 2
      ;;
    -s|--snapshot-id)
      SNAPSHOT_ID="$2"
      shift 2
      ;;
    -z|--snapshot-zone)
      SNAPSHOT_ZONE="$2"
      shift 2
      ;;
    -a|--snapshot-api)
      SNAPSHOT_API="$2"
      shift 2
      ;;
    -p|--project-id)
      PROJECT_ID="$2"
      shift 2
      ;;
    -c|--arch)
      ARCH="$2"
      shift 2
      ;;
    -h|--help)
      print_help
      exit 0
      ;;
    *)
      echo "Unknown option: $1"
      print_help
      exit 1
      ;;
  esac
done

if [[ -z "$IMAGE_NAME" ]]; then
  echo "No image name specified"
  exit 1
fi

# Export snapshot to S3.

if [[ -z "$BUCKET_NAME" ]]; then
  echo "No bucket name specified"
  exit 1
fi

if [[ -z "$SNAPSHOT_ID" ]]; then
  echo "No snapshot ID specified"
  exit 1
fi

if [[ -z "$SNAPSHOT_ZONE" ]]; then
  echo "No snapshot zone specified"
  exit 1
fi

echo "Uploading snapshot to S3"

case "$SNAPSHOT_API" in

  block)
    echo "   ... Exporting block snapshot to bucket"
    scw block snapshot export-to-object-storage snapshot-id="${SNAPSHOT_ID}" bucket="${BUCKET_NAME}" key="${IMAGE_NAME}" zone="${SNAPSHOT_ZONE}" > /dev/null
    echo "   ... Waiting for export to finish"
    scw block snapshot wait "${SNAPSHOT_ID}" zone="${SNAPSHOT_ZONE}" > /dev/null
    ;;

  instance)
    echo "   ... Exporting instance snapshot to bucket"
    scw instance snapshot export snapshot-id="${SNAPSHOT_ID}" bucket="${BUCKET_NAME}" key="${IMAGE_NAME}" zone="${SNAPSHOT_ZONE}" > /dev/null
    echo "   ... Waiting for export to finish"
    scw instance snapshot wait "${SNAPSHOT_ID}" zone="${SNAPSHOT_ZONE}" > /dev/null
    ;;

  *)
    echo "unknown snapshot API"
    exit 1
    ;;
esac

# Import snapshot from S3

export IFS=","
for target_zone in ${TARGET_ZONES}; do

  echo "Importing snapshot to ${target_zone} from bucket"

  case "$TARGET_API" in

    block)
      echo "   ... Importing block snapshot to ${target_zone} from bucket"
      imported_snapshot_id=$(scw block snapshot import-from-object-storage bucket="${BUCKET_NAME}" key="${IMAGE_NAME}" name="${IMAGE_NAME}" project-id="${PROJECT_ID}" zone="${target_zone}" -o template="{{  .ID }}")
      echo "   ... Waiting for import to finish"
      scw block snapshot wait "${imported_snapshot_id}" zone="${target_zone}" > /dev/null
      ;;

    instance)
      echo "   ... Importing instance snapshot to ${target_zone} from bucket"
      imported_snapshot_id=$(scw instance snapshot create bucket="${BUCKET_NAME}" key="${IMAGE_NAME}" name="${IMAGE_NAME}" project-id="${PROJECT_ID}" zone="${target_zone}" volume-type=l_ssd -o template="{{  .Snapshot.ID }}")
      echo "   ... Waiting for import to finish"
      scw instance snapshot wait "${imported_snapshot_id}" zone="${target_zone}" > /dev/null
      ;;

    *)
      echo "unknown snapshot API"
      exit 1
      ;;
  esac

  echo "   ... Creating instance image in ${target_zone}"
  scw instance image create name="${IMAGE_NAME}" snapshot-id="${imported_snapshot_id}" project-id="${PROJECT_ID}" arch="${ARCH}" zone="${target_zone}" > /dev/null
done

echo "Images were successfully created. The S3 bucket needs to be cleaned manually."
