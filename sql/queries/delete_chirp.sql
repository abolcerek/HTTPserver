-- name: DeleteChirp :exec
DELETE FROM chirps 
WHERE id = $1 and user_id = $2;