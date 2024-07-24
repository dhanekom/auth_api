TRUNCATE TABLE public.users;
TRUNCATE TABLE public.verification;

INSERT INTO public.users (email,"password",is_verified,created_at,updated_at) VALUES
	 ('unverified@gmail.com','$2a$14$efGcxhO6bZZ/j36eglsix.m4gzy94PQ.FceZUOQLVX.knBODFKLnK',false,'2024-07-23 13:33:08.951815','2024-07-23 13:33:08.951815');
INSERT INTO public.users (email,"password",is_verified,created_at,updated_at) VALUES
	 ('verified@gmail.com','$2a$14$efGcxhO6bZZ/j36eglsix.m4gzy94PQ.FceZUOQLVX.knBODFKLnK',true,'2024-07-23 13:33:08.951815','2024-07-23 13:33:08.951815');
INSERT INTO public.users (email,"password",is_verified,created_at,updated_at) VALUES
	 ('retries@gmail.com','$2a$14$efGcxhO6bZZ/j36eglsix.m4gzy94PQ.FceZUOQLVX.knBODFKLnK',false,'2024-07-23 13:33:08.951815','2024-07-23 13:33:08.951815');

INSERT INTO public.verification (email,verification_code,expires_at,attempts_remaining,created_at,updated_at) VALUES
	 ('unverified@gmail.com','NGBGVM','2024-07-24 15:33:36.106086',3,'2024-07-23 13:33:36.107427','2024-07-23 13:33:36.107427');
INSERT INTO public.verification (email,verification_code,expires_at,attempts_remaining,created_at,updated_at) VALUES
	 ('retries@gmail.com','NGBGVM','2024-07-24 15:33:36.106086',0,'2024-07-23 13:33:36.107427','2024-07-23 13:33:36.107427');
