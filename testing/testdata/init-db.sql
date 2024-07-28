TRUNCATE TABLE public.users;
TRUNCATE TABLE public.verification;

INSERT INTO public.users (user_id, email,"password",is_verified,created_at,updated_at) VALUES
	 ('7b8c7b8f-b2d7-4045-af58-a49db6d47a81', 'unverified@gmail.com','$2a$14$efGcxhO6bZZ/j36eglsix.m4gzy94PQ.FceZUOQLVX.knBODFKLnK',false,'2024-07-23 13:33:08.951815','2024-07-23 13:33:08.951815');
INSERT INTO public.users (user_id, email,"password",is_verified,created_at,updated_at) VALUES
	 ('74a8ebde-489d-4c04-843b-8f22f19bae0b', 'verified@gmail.com','validpass',true,'2024-07-23 13:33:08.951815','2024-07-23 13:33:08.951815');
INSERT INTO public.users (user_id, email,"password",is_verified,created_at,updated_at) VALUES
	 ('809c4dbd-7363-4328-9ebb-e19f899681d3', 'invalidpassword@gmail.com','blah',true,'2024-07-23 13:33:08.951815','2024-07-23 13:33:08.951815');
INSERT INTO public.users (user_id, email,"password",is_verified,created_at,updated_at) VALUES
	 ('b8317d24-87ff-4854-b6eb-90f554877ee7', 'toomanyattempts@gmail.com','$2a$14$efGcxhO6bZZ/j36eglsix.m4gzy94PQ.FceZUOQLVX.knBODFKLnK',false,'2024-07-23 13:33:08.951815','2024-07-23 13:33:08.951815');
INSERT INTO public.users (user_id, email,"password",is_verified,created_at,updated_at) VALUES
	 ('ff2db9c1-0ef7-400d-9642-e0613f1282bb', 'noverification@gmail.com','$2a$14$efGcxhO6bZZ/j36eglsix.m4gzy94PQ.FceZUOQLVX.knBODFKLnK',false,'2024-07-23 13:33:08.951815','2024-07-23 13:33:08.951815');
INSERT INTO public.users (user_id, email,"password",is_verified,created_at,updated_at) VALUES
	 ('e178843e-cc39-435d-abc6-8ee94b9f3134', 'expiredverification@gmail.com','$2a$14$efGcxhO6bZZ/j36eglsix.m4gzy94PQ.FceZUOQLVX.knBODFKLnK',false,'2024-07-23 13:33:08.951815','2024-07-23 13:33:08.951815');
INSERT INTO public.users (user_id, email,"password",is_verified,created_at,updated_at) VALUES
	 ('0460d39a-9c81-48bd-86ed-7154f44ac617', 'authcodefailed@gmail.com','validpass',true,'2024-07-23 13:33:08.951815','2024-07-23 13:33:08.951815');

INSERT INTO public.verification (email,verification_code,expires_at,attempts_remaining,created_at,updated_at) VALUES
	 ('unverified@gmail.com','ABCDEF','2099-07-24 15:33:36.106086',3,'2024-07-23 13:33:36.107427','2024-07-23 13:33:36.107427');
INSERT INTO public.verification (email,verification_code,expires_at,attempts_remaining,created_at,updated_at) VALUES
	 ('toomanyattempts@gmail.com','ABCDEF','2099-07-24 15:33:36.106086',0,'2024-07-23 13:33:36.107427','2024-07-23 13:33:36.107427');
INSERT INTO public.verification (email,verification_code,expires_at,attempts_remaining,created_at,updated_at) VALUES
	 ('expiredverification@gmail.com','ABCDEF','2000-07-24 15:33:36.106086',0,'2024-07-23 13:33:36.107427','2024-07-23 13:33:36.107427');
