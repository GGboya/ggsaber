-- 创建数据库
CREATE DATABASE IF NOT EXISTS saber CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

-- 创建用户
CREATE USER IF NOT EXISTS 'saber'@'localhost' IDENTIFIED BY 'saber123';

-- 授权
GRANT ALL PRIVILEGES ON saber.* TO 'saber'@'localhost';
FLUSH PRIVILEGES;

-- 使用数据库
USE saber; 