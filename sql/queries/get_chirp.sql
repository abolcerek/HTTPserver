-- name: GetChirp :one
SELECT * FROM chirps
WHERE id = $1 LIMIT 1;