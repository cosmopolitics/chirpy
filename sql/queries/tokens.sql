-- name: AddUsersRefreshToken :one
insert into 
  refresh_tokens (
    token, created_at, updated_at, user_id, expired_at
  )
values (
  ($1),
  (select now()),
  (select now()),
  $2,
  $3
) returning $1;
--

-- name: RevokeRT :exec
update 
  refresh_tokens
set
  updated_at = (select now()),
  revoked_at = (select now())
where
  token = $1;
--

-- name: GetUserByRT :one
select 
  u.*, 
  refresh_tokens.revoked_at 
from 
  users u
inner join 
  refresh_tokens
  on refresh_tokens.user_id = u.id
where 
  refresh_tokens.token = $1;
--

