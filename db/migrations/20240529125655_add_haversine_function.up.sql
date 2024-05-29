CREATE OR REPLACE FUNCTION haversine(lat1 float8, lon1 float8, lat2 float8, lon2 float8)
RETURNS float8 AS $$
DECLARE
    r float8 := 6371; -- Earth radius in kilometers
    dlat float8 := radians(lat2 - lat1);
    dlon float8 := radians(lon2 - lon1);
    a float8 := sin(dlat / 2) ^ 2 + cos(radians(lat1)) * cos(radians(lat2)) * sin(dlon / 2) ^ 2;
    c float8 := 2 * atan2(sqrt(a), sqrt(1 - a));
BEGIN
    RETURN r * c;
END;
$$ LANGUAGE plpgsql IMMUTABLE STRICT;
