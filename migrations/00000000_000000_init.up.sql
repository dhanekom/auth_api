CREATE TABLE if not exists public.users (
  user_id uuid PRIMARY KEY,
  email varchar(255) not null,
  password text not null,
  is_verified bool not null DEFAULT false,
  created_at TIMESTAMP not null DEFAULT now(),
  updated_at TIMESTAMP not null DEFAULT now(),
  UNIQUE(email)
);

CREATE INDEX if not exists idx_users_email ON users(email);

CREATE TABLE if not exists public.verification (
  email varchar(255) PRIMARY KEY,
  verification_code varchar(255) not null,
  expires_at TIMESTAMP not null,
  attempts_remaining int not null,
  created_at TIMESTAMP not null DEFAULT now(),
  updated_at TIMESTAMP not null DEFAULT now()
);

CREATE INDEX if not exists idx_verification_email ON verification(email);