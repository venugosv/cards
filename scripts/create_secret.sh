#!/bin/bash
set -e

print_usage () {
  echo "Usage: DEPLOY_ENV=[st|sit|sit-n|intpnv|pnv|preprod|preprod-k|prod] SERVICE=cards ./scripts/create_secret.sh {SecretName} --from-literal={key}={value}  --from-literal={key}={value} ...

Example:
    DEPLOY_ENV=st SERVICE=cards ./scripts/create_secret.sh apic-ids --from-literal=k1=v1  --from-literal=k2=v2
"
}

if [[ $# -lt 2 ]]; then
    echo "Illegal number of parameters"
    print_usage
    exit 1
fi

if [[ -z "${SERVICE}" ]]; then
  echo "SERVICE must be specified: one of 'cards', 'cardcontrols'"
  print_usage
  exit 1
fi

if [[ -z "${DEPLOY_ENV}" ]]; then
  echo "DEPLOY_ENV must be specified: one of 'st', 'sit','sit-n', 'intpnv', 'pnv', 'preprod', 'preprod-k', 'prod'"
  print_usage
  exit 1
fi

create_secret () {
  # Download public key for encrypting secrets
  if [[ "${DEPLOY_ENV}" == "prod" ]]; then
    FILE=/tmp/public-key-prod.pem
    URL=https://sealed-secrets.kube-system.apps.x.gcp.anz/v1/cert.pem
  else
    FILE=/tmp/public-key-np.pem
    URL=https://sealed-secrets.kube-system.apps.x.gcpnp.anz/v1/cert.pem
  fi

  if test -f "$FILE"; then
    echo "$FILE exists."
  else
    curl -k $URL > $FILE
  fi

  echo ./config/${SERVICE}/secrets/${DEPLOY_ENV}/$1-sealed-secret.yaml

  kubectl create secret generic "fabric-${SERVICE}-$1" --namespace "fabric-services-cde-${DEPLOY_ENV}" "${@:2}"\
    --dry-run -o yaml | kubeseal --format yaml --cert $FILE > \
    ./config/${SERVICE}/config/secrets/${DEPLOY_ENV}/$1-sealed-secret.yaml

  echo "secret ./config/${SERVICE}/config/secrets/${DEPLOY_ENV}/$1-sealed-secret.yaml created!"
}

echo "Creating secret for ${SERVICE} in ${DEPLOY_ENV}..."

create_secret $@

echo "Done!"
