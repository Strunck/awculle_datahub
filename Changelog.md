# Changelog

## Release v0.0.27
- Handling Symbolconfiguration 
- Better Error in case of CreateMonitoredItem fails
- Subcribe in batches of 100
- settings.json filename for Symbolconfiguration


## Release v0.0.28
- Fix ConvertSymbol
- add value nil check
- Better error

## Release v0.0.29
- If EMeter is not in the PLC, programm just ignores it
- avoid panic, when type casting fails in compose.go

## Release v0.0.30
- Extend Report api, so it is more like uv3
- Add preperation for weekly report

# Release v0.0.31
- Add Saison Report from H.H.Pohl (EOA) spec

# Release v0.0.32
- Unix Timestamp instead of partial RFC3339

# Release v0.0.33
- Fix wrong Tauwasser scale with hacky solution
- Rearange Report functionallity to improve seperation of concerns
- Ort Report can have start and end date
- Historical Tagesreport
- Fix race condition

# Release v0.0.34 (TODO)
- Ort-Reporte: 
    - O2/CO2 have 2 digits fixed
    - "Takte" no digits
- Aussentemperatur for wochenreport, zweiwochenreport
- zweiwochendiagramm
- Tagesreport: 
    - Tauwasser-Zuwachs 1-digit /10
- Fix race conditions in DailyReportGaterer and RecieveOpcUaMessage -> storeAndPublish

TODO:
Wochenreport:
    - Sorte
    - Einlagerumgsmenge
    - Einlagerungsdatum
Saisonreport
    - Datum als dd.mm.yyyyy
Möglichkeit für Kundenlogo
