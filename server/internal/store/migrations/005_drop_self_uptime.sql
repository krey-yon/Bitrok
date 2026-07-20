-- A process-local health poll cannot observe its own downtime and therefore
-- must not be presented as service availability data.
DROP TABLE IF EXISTS uptime_checks;
