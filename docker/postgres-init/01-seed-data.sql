-- Seed data for banterbus application
-- This script populates the database with initial groups and questions

-- Insert groups
INSERT INTO questions_groups (id, group_name, group_type) VALUES
    ('01945c66-891a-7894-ae92-c18087c73a23'::uuid, 'programming', 'questions'),
    ('01945c66-891c-7942-9a2a-339a62a74800'::uuid, 'horse', 'questions'),
    ('01945c66-891c-7aa2-b6ca-088679706a5b'::uuid, 'colour', 'questions'),
    ('01945c66-891b-7d3e-804c-f2e170b0b0ce'::uuid, 'cat', 'questions'),
    ('01945c66-891c-74d5-9870-7a8777e37588'::uuid, 'bike', 'questions'),
    ('01945c66-891c-7d8a-b404-be384c9515a6'::uuid, 'animal', 'questions'),
    ('01947acd-d953-76d1-881b-247a59906035'::uuid, 'person', 'questions'),
    ('01947acd-d953-76d1-881b-247a59906036'::uuid, 'food', 'questions')
ON CONFLICT (id) DO NOTHING;

-- Insert questions
INSERT INTO questions (id, game_name, round_type, group_id) VALUES
    ('4b1355bb-82de-40c8-8eda-0c634091cc3c'::uuid, 'fibbing_it', 'most_likely', '01947acd-d953-76d1-881b-247a59906035'::uuid),
    ('a91af98c-f989-4e00-aa14-7a34e732519e'::uuid, 'fibbing_it', 'most_likely', '01947acd-d953-76d1-881b-247a59906035'::uuid),
    ('fac6a98f-e3b5-4328-999c-b39fd86657ba'::uuid, 'fibbing_it', 'most_likely', '01947acd-d953-76d1-881b-247a59906035'::uuid),
    ('6b60f097-b714-4f9e-b8cb-de75a7890381'::uuid, 'fibbing_it', 'most_likely', '01945c66-891c-7942-9a2a-339a62a74800'::uuid),
    ('6b60f097-b714-4f9e-b8cb-de75a7890382'::uuid, 'fibbing_it', 'most_likely', '01945c66-891c-7942-9a2a-339a62a74800'::uuid),
    ('93dd56a8-c8a3-4c63-93dc-9d890c4d2b74'::uuid, 'fibbing_it', 'free_form', '01945c66-891a-7894-ae92-c18087c73a23'::uuid),
    ('066e7a8a-b0b7-44d4-b882-582a64151c15'::uuid, 'fibbing_it', 'free_form', '01945c66-891a-7894-ae92-c18087c73a23'::uuid),
    ('654327b9-36a2-4d75-b4bf-d68d19fcfe7c'::uuid, 'fibbing_it', 'free_form', '01945c66-891a-7894-ae92-c18087c73a23'::uuid),
    ('281bc3c7-f55d-4a8a-88cf-4e0d67d2825e'::uuid, 'fibbing_it', 'free_form', '01945c66-891b-7d3e-804c-f2e170b0b0ce'::uuid),
    ('fc1a3c9f-3d98-452e-b77e-c6c7f353176d'::uuid, 'fibbing_it', 'free_form', '01945c66-891b-7d3e-804c-f2e170b0b0ce'::uuid),
    ('393dae17-84fe-449d-ba0f-8c9d320a46e6'::uuid, 'fibbing_it', 'free_form', '01945c66-891b-7d3e-804c-f2e170b0b0ce'::uuid),
    ('393dae17-84fe-449d-ba0f-8c9d320a46e7'::uuid, 'fibbing_it', 'free_form', '01945c66-891b-7d3e-804c-f2e170b0b0ce'::uuid),
    ('8aa9f87f-31d9-4421-aae5-2024ca730348'::uuid, 'fibbing_it', 'free_form', '01945c66-891c-74d5-9870-7a8777e37588'::uuid),
    ('8aa9f87f-31d9-4421-aae5-2024ca730350'::uuid, 'fibbing_it', 'free_form', '01945c66-891c-74d5-9870-7a8777e37588'::uuid),
    ('8aa9f87f-31d9-4421-aae5-2024ca730351'::uuid, 'fibbing_it', 'free_form', '01945c66-891c-74d5-9870-7a8777e37588'::uuid),
    ('8aa9f87f-31d9-4421-aae5-2024ca730352'::uuid, 'fibbing_it', 'free_form', '01945c66-891c-74d5-9870-7a8777e37588'::uuid),
    ('8aa9f87f-31d9-4421-aae5-2024ca730353'::uuid, 'fibbing_it', 'free_form', '01947acd-d953-76d1-881b-247a59906036'::uuid),
    ('8aa9f87f-31d9-4421-aae5-2024ca730354'::uuid, 'fibbing_it', 'free_form', '01947acd-d953-76d1-881b-247a59906036'::uuid),
    ('89b20c84-12ae-444d-ad9c-26f72d3f28ab'::uuid, 'fibbing_it', 'multiple_choice', '01945c66-891c-7942-9a2a-339a62a74800'::uuid),
    ('68ed9133-dc58-41bb-b642-c48470998127'::uuid, 'fibbing_it', 'multiple_choice', '01945c66-891c-7942-9a2a-339a62a74800'::uuid),
    ('e90d613d-2e6c-4331-9204-9b685c0795b7'::uuid, 'fibbing_it', 'multiple_choice', '01945c66-891c-7d8a-b404-be384c9515a6'::uuid),
    ('89deb03f-66be-4265-91e6-dedd9227718a'::uuid, 'fibbing_it', 'multiple_choice', '01945c66-891c-7d8a-b404-be384c9515a6'::uuid)
ON CONFLICT (id) DO NOTHING;

-- Insert question translations
INSERT INTO questions_i18n (id, question, locale, question_id) VALUES
    (gen_random_uuid(), 'to get arrested', 'en-GB', '4b1355bb-82de-40c8-8eda-0c634091cc3c'::uuid),
    (gen_random_uuid(), 'to eat ice-cream from the tub', 'en-GB', 'a91af98c-f989-4e00-aa14-7a34e732519e'::uuid),
    (gen_random_uuid(), 'to fight a police officer', 'en-GB', 'fac6a98f-e3b5-4328-999c-b39fd86657ba'::uuid),
    (gen_random_uuid(), 'to steal a horse', 'en-GB', '6b60f097-b714-4f9e-b8cb-de75a7890381'::uuid),
    (gen_random_uuid(), 'to ride a horse', 'en-GB', '6b60f097-b714-4f9e-b8cb-de75a7890382'::uuid),
    (gen_random_uuid(), 'What do you think about programmers', 'en-GB', '93dd56a8-c8a3-4c63-93dc-9d890c4d2b74'::uuid),
    (gen_random_uuid(), 'What don''t you like about programmers', 'en-GB', '066e7a8a-b0b7-44d4-b882-582a64151c15'::uuid),
    (gen_random_uuid(), 'what don''t you think about programmers', 'en-GB', '654327b9-36a2-4d75-b4bf-d68d19fcfe7c'::uuid),
    (gen_random_uuid(), 'what dont you think about cats', 'en-GB', '281bc3c7-f55d-4a8a-88cf-4e0d67d2825e'::uuid),
    (gen_random_uuid(), 'what don''t you like about cats', 'en-GB', 'fc1a3c9f-3d98-452e-b77e-c6c7f353176d'::uuid),
    (gen_random_uuid(), 'what do you like about cats', 'en-GB', '393dae17-84fe-449d-ba0f-8c9d320a46e6'::uuid),
    (gen_random_uuid(), 'what do you think about cats', 'en-GB', '393dae17-84fe-449d-ba0f-8c9d320a46e7'::uuid),
    (gen_random_uuid(), 'Favourite bike colour', 'en-GB', '8aa9f87f-31d9-4421-aae5-2024ca730348'::uuid),
    (gen_random_uuid(), 'Who would win in a fight a bike or a car', 'en-GB', '8aa9f87f-31d9-4421-aae5-2024ca730350'::uuid),
    (gen_random_uuid(), 'What color bike do you prefer', 'en-GB', '8aa9f87f-31d9-4421-aae5-2024ca730351'::uuid),
    (gen_random_uuid(), 'How fast can a bike go', 'en-GB', '8aa9f87f-31d9-4421-aae5-2024ca730352'::uuid),
    (gen_random_uuid(), 'What is your favorite food', 'en-GB', '8aa9f87f-31d9-4421-aae5-2024ca730353'::uuid),
    (gen_random_uuid(), 'What food do you dislike', 'en-GB', '8aa9f87f-31d9-4421-aae5-2024ca730354'::uuid),
    (gen_random_uuid(), 'What do you think about camels', 'en-GB', '89b20c84-12ae-444d-ad9c-26f72d3f28ab'::uuid),
    (gen_random_uuid(), 'What do you think about horses', 'en-GB', '68ed9133-dc58-41bb-b642-c48470998127'::uuid),
    (gen_random_uuid(), 'Are cats cute', 'en-GB', 'e90d613d-2e6c-4331-9204-9b685c0795b7'::uuid),
    (gen_random_uuid(), 'Dogs are cuter than cats', 'en-GB', '89deb03f-66be-4265-91e6-dedd9227718a'::uuid)
ON CONFLICT (question, locale) DO NOTHING;