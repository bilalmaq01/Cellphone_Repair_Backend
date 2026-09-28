-- Device condition checklist captured at intake, stored as a JSON object of
-- known item keys to yes/no answers (e.g. {"powers_on": true}). Unanswered
-- items are simply absent from the object.
alter table repairs add column if not exists intake_checklist jsonb;