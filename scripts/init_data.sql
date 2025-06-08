-- 创建数据库
CREATE DATABASE IF NOT EXISTS saber CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE saber;

-- 插入示例用户
INSERT IGNORE INTO users (username, email, password, rating, wins, losses, created_at, updated_at) VALUES
('alice', 'alice@example.com', '5d41402abc4b2a76b9719d911017c592', 1200, 5, 3, NOW(), NOW()),
('bob', 'bob@example.com', '5d41402abc4b2a76b9719d911017c592', 1150, 3, 4, NOW(), NOW()),
('charlie', 'charlie@example.com', '5d41402abc4b2a76b9719d911017c592', 1300, 8, 2, NOW(), NOW()),
('david', 'david@example.com', '5d41402abc4b2a76b9719d911017c592', 1100, 2, 5, NOW(), NOW());

-- 插入示例题目
INSERT IGNORE INTO problems (title, description, difficulty, time_limit, memory_limit, test_cases, created_at, updated_at) VALUES
('两数之和', 
'给定一个整数数组 nums 和一个整数目标值 target，请你在该数组中找出和为目标值 target 的那两个整数，并返回它们的数组下标。

示例：
输入：nums = [2,7,11,15], target = 9
输出：[0,1]',
'easy', 1000, 256,
'[{"input":"4\\n2 7 11 15\\n9", "output":"0 1"}, {"input":"3\\n3 2 4\\n6", "output":"1 2"}]',
NOW(), NOW()),

('反转链表', 
'给你单链表的头节点 head ，请你反转链表，并返回反转后的链表。

示例：
输入：head = [1,2,3,4,5]
输出：[5,4,3,2,1]',
'medium', 1000, 256,
'[{"input":"5\\n1 2 3 4 5", "output":"5 4 3 2 1"}, {"input":"2\\n1 2", "output":"2 1"}]',
NOW(), NOW());

-- 设置用户密码为 "hello"（MD5加密）
-- 5d41402abc4b2a76b9719d911017c592 是 "hello" 的MD5值 