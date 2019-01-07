-- noinspection SqlNoDataSourceInspectionForFile

-- name: search-bar
SELECT "Name", weight, height
  FROM pokemon
    WHERE "Name"
      LIKE $1

-- name: flavor-text
SELECT flavor_text, version_id
  FROM pokemon
    INNER JOIN flavor_text
      ON pokemon.species_id = flavor_text.species_id
    WHERE "Name"=$1
      AND language_id=9
      AND version_id > 1
