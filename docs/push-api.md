# Mapon Push API Documentation

## Overview

The Push API V1 sends data packets sequentially to configured endpoints. One pack is transmitted at a time. Messages must receive acknowledgement before the next pack transmits, or they queue for retry. The system maintains a 12-hour TTL and 100k maximum queue size per endpoint.

## Core Requirements

- **Acknowledgement format:** `{"status":"ok"}` — required for Mapon to send the next pack
- **Data pack types constraint:** Cannot mix within a single endpoint scope
- **Boolean values:** Use 0 (off) and 1 (on)
- **Pack delivery:** Sequential, one at a time per endpoint
- **Retry policy:** Failed deliveries queue for up to 12 hours, max 100k elements per endpoint

## Request Structure

A typical pack includes:
- `id` — Mapon push event ID
- `car_id` — Mapon unit/car ID
- `pack_id` — Pack type identifier
- `device_id` — Device ID
- `company_id` — Company ID
- `gmt` — DateTime in format `YYYY-MM-DD HH:MM:SS` (UTC)
- Pack-specific fields (e.g., latitude, longitude, speed for position data)

## Car Pack Types

The system includes 40+ car-related event types:

### Basic Vehicle Data
- **#1** Position — GPS position, speed, heading, altitude
- **#2** Switch states
- **#3** Ignition — Ignition on/off
- **#4** External power
- **#5** Fuel — Fuel level

### Reefer Units
- **#8** Configuration
- **#9** Mode
- **#10** Compartment
- **#11** Temperature
- **#12** Hours
- **#13** Voltage
- **#14** Alarms

### Vehicle Events
- **#15** OBD events
- **#16-17** Vehicle updated/created
- **#18** Device removal
- **#19** Driver DDD downloads
- **#20** Vehicle DDD downloads

### Monitoring
- **#21** Zone entry/exit
- **#22** Behavior events
- **#23** SOS button
- **#24** Relay changes
- **#25** BLE beacons
- **#26** Odometer
- **#27-28** Crash detection

### CAN Data
- **#30-31** RPM average/max
- **#32** Fuel level (CAN)
- **#33-34** Service/total distance
- **#35** Total fuel
- **#36** Engine hours
- **#37** Ambient temperature
- **#38** Axle weight
- **#39** Trailer load
- **#54** PTO switch
- **#55** Temperature sensors

### Additional Pack Categories
- **#101** Device — Device configuration updates
- **#201** Company — Company information changes
- **#301-303** User — User lifecycle events
- **#401** 3rd Party — Custom field updates
- **#501-504** Route Planning — PTA/ETA changes, status updates, attachments
- **#601-602** BLE Tags — Tag scanning and location tracking

## Supported Car Packs for Initial Integration

Phase 1 focuses on car packs only, excluding CAN packs (#30-37) and non-car packs:
- **#1** Position
- **#3** Ignition
- **#5** Fuel
- **#26** Odometer
- **#55** Temperature sensors

## Field Types and Units

All car pack data uses OEM-native units:
- **Position:** latitude/longitude (degrees), speed (km/h), heading (degrees 0-360), altitude (meters)
- **Fuel:** liters
- **Odometer:** meters (or convert from km as needed)
- **Temperature:** Celsius
- **Ignition:** boolean (0=off, 1=on)

