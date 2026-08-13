DROP TRIGGER IF EXISTS update_user_twofa_updated_at ON user_twofa;
DROP TRIGGER IF EXISTS update_project_applications_updated_at ON project_applications;
DROP TRIGGER IF EXISTS update_vacancy_responses_updated_at ON vacancy_responses;
DROP TRIGGER IF EXISTS update_vacancies_updated_at ON vacancies;
DROP TRIGGER IF EXISTS update_projects_updated_at ON projects;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP TABLE IF EXISTS notifications;
DROP TABLE IF EXISTS user_twofa;
DROP TABLE IF EXISTS user_devices;
DROP TABLE IF EXISTS messages;
DROP TABLE IF EXISTS vacancy_responses;
DROP TABLE IF EXISTS vacancies;
DROP TABLE IF EXISTS project_applications;
DROP TABLE IF EXISTS project_members;
DROP TABLE IF EXISTS projects;