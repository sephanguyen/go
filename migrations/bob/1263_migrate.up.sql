UPDATE usr_email AS um

SET email = lower(u.email),
    updated_at = now()

FROM users AS u

WHERE um.usr_id = u.user_id AND
      um.email != lower(u.email)
