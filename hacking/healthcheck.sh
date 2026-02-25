#!/usr/bin/env bash
set -euo pipefail

# Konfigurierbare Parameter
URL="${URL:-https://localhost/middleware/healthcheck}"
SYSTEMD_SERVICE="${SYSTEMD_SERVICE:-container-middleware.service}"
CURL_OPTS=(
  --silent            # keine Fortschrittsausgabe
  --show-error        # Fehler anzeigen
  --fail-with-body    # non-2xx als Fehler behandeln
  --max-time 10       # Timeout 10s
  --insecure          # falls selbstsigniertes Zertifikat, kann per INSECURE=1 gesetzt werden
)

# # Falls selbstsigniertes Zertifikat: INSECURE=1 setzen (z.B. per Environment)
# if [[ "${INSECURE:-0}" == "1" ]]; then
#   CURL_OPTS+=( --insecure )
# fi

# Optional: spezifisches CA-Zertifikat nutzen
# Beispiel: export CURL_CA_BUNDLE=/etc/ssl/certs/your-ca.pem
if [[ -n "${CURL_CA_BUNDLE:-}" ]]; then
  CURL_OPTS+=( --cacert "$CURL_CA_BUNDLE" )
fi

log() {
  # nutzt systemd-journald falls verfügbar
  logger -t middleware-healthcheck "$*"
}

restart_service() {
  log "Healthcheck fehlgeschlagen oder Ok=false. Starte ${SYSTEMD_SERVICE} neu…"
  if ! systemctl --user restart "$SYSTEMD_SERVICE"; then
    log "FEHLER: Neustart von ${SYSTEMD_SERVICE} fehlgeschlagen."
    exit 1
  fi
  log "Neustart von ${SYSTEMD_SERVICE} abgeschlossen."
}

main() {
  # Endpoint abrufen
  if ! RESPONSE="$(curl "${CURL_OPTS[@]}" "$URL")"; then
    log "FEHLER: curl-Aufruf fehlgeschlagen. Triggere Neustart."
    restart_service
    exit 0
  fi

  # JSON validieren
  if ! echo "$RESPONSE" | jq -e . >/dev/null 2>&1; then
    log "FEHLER: Ungültiges JSON erhalten: $RESPONSE"
    restart_service
    exit 0
  fi

  # Ok-Feld extrahieren; standardmäßig false, falls nicht vorhanden
  OK_VALUE="$(echo "$RESPONSE" | jq -r 'if has("Ok") then .Ok else false end')"

  # jq gibt true/false ohne Quotes aus; String-Absicherung:
  case "$OK_VALUE" in
    true|True|TRUE)
      log "Healthcheck OK (Ok=true)."
      ;;
    *)
      log "Healthcheck meldet Ok=false. Triggere Neustart."
      restart_service
      ;;
  esac
}

main