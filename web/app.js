// 全局变量
let currentUser = null;
let codeEditor = null;
let ws = null;
let battleId = null;
let battleTimer = null;
let startTime = null;

// 核心模式相关变量
let currentProblem = null;
let codeTemplates = null;
let functionSignature = null;
let coreTestCases = null;

// API基地址
const API_BASE = 'http://localhost:8080/api/v1';

// 调试信息
console.log('Go Saber JavaScript loaded');

// 页面加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    console.log('DOM loaded, initializing app...');
    initializeApp();
});

// 确保在页面完全加载后也初始化
window.addEventListener('load', function() {
    console.log('Window loaded, checking initialization...');
    if (!currentUser && document.getElementById('loginForm').classList.contains('hidden')) {
        console.log('Forcing login display...');
        showLogin();
    }
});

// 初始化应用
function initializeApp() {
    console.log('Initializing app...');
    
    // 检查是否有保存的用户信息
    const savedUser = localStorage.getItem('currentUser');
    if (savedUser) {
        try {
            currentUser = JSON.parse(savedUser);
            console.log('Found saved user:', currentUser.username);
            
            // 建立WebSocket连接
            connectWebSocket();
            
            showMainContent();
        } catch (error) {
            console.error('Error parsing saved user:', error);
            localStorage.removeItem('currentUser');
            showLogin();
        }
    } else {
        console.log('No saved user, showing login');
        showLogin();
    }
    
    // 初始化代码编辑器 (延迟执行)
    setTimeout(initCodeEditor, 1000);
}

// 初始化代码编辑器
function initCodeEditor() {
    console.log('Initializing Monaco Editor...');
    
    // 确保Monaco Editor资源已加载
    if (typeof monaco !== 'undefined') {
        initMonacoEditor();
        return;
    }
    
    // 动态加载Monaco Editor
    window.require.config({ 
        paths: { 
            'vs': 'https://cdnjs.cloudflare.com/ajax/libs/monaco-editor/0.44.0/min/vs' 
        }
    });
    
    window.require(['vs/editor/editor.main'], function () {
        console.log('Monaco Editor loaded successfully');
        initMonacoEditor();
    });
}

// 初始化Monaco编辑器
function initMonacoEditor() {
    const container = document.getElementById('codeEditor');
    if (!container) {
        console.error('Code editor container not found');
        return;
    }
    
    // 清空容器
    container.innerHTML = '';
    
    try {
        codeEditor = monaco.editor.create(container, {
            value: 'package main\n\nimport "fmt"\n\nfunc main() {\n    // Your code here\n    m := map[int]int{}\n    fmt.Println(m)\n}',
            language: 'go',
            theme: 'vs-dark',
            automaticLayout: true,
            fontSize: 14,
            fontFamily: '"Consolas", "Monaco", monospace',
            lineNumbers: 'on',
            minimap: { enabled: false },
            scrollBeyondLastLine: false,
            wordWrap: 'off',
            readOnly: false,
            contextmenu: false,
            selectOnLineNumbers: true,
            cursorStyle: 'line',
            cursorBlinking: 'blink',
            renderWhitespace: 'none',
            renderLineHighlight: 'line',
            
            // 基本编辑功能
            autoClosingBrackets: 'always',
            autoClosingQuotes: 'always',
            tabSize: 4,
            insertSpaces: true,
            
            // 禁用复杂功能
            quickSuggestions: false,
            suggestOnTriggerCharacters: false,
            acceptSuggestionOnEnter: 'off',
            wordBasedSuggestions: false,
            folding: false,
            bracketPairColorization: { enabled: false },
            guides: {
                bracketPairs: false,
                indentation: false
            },
            
            // 滚动条简化
            scrollbar: {
                vertical: 'visible',
                horizontal: 'visible'
            }
        });
        
        console.log('Monaco Editor initialized successfully');
        
        // 添加自定义键盘快捷键
        addCustomKeybindings();
        
        // 添加自定义命令
        addCustomCommands();
        
        // 确保编辑器获得焦点
        setTimeout(() => {
            codeEditor.focus();
        }, 100);
        
        // 监听编辑器变化
        codeEditor.onDidChangeModelContent(function() {
            console.log('Code content changed');
        });
        
        // 语言切换事件
        const languageSelect = document.getElementById('languageSelect');
        if (languageSelect) {
            languageSelect.addEventListener('change', function() {
                const language = this.value;
                updateEditorLanguageAndTemplate(language);
            });
        }
    } catch (error) {
        console.error('Error initializing Monaco Editor:', error);
        // 回退到简单的textarea
        container.innerHTML = '<textarea id="fallbackEditor" style="width:100%;height:400px;font-family:monospace;"></textarea>';
    }
}

// 添加基本键盘导航快捷键
function addCustomKeybindings() {
    if (!codeEditor) return;
    
    // 确保基本导航快捷键正常工作
    // Home: 跳转到行首 (Monaco默认支持)
    // End: 跳转到行尾 (Monaco默认支持)
    // Ctrl+Home: 跳转到文档开始
    codeEditor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.Home, function() {
        codeEditor.setPosition({ lineNumber: 1, column: 1 });
    });
    
    // Ctrl+End: 跳转到文档结尾
    codeEditor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.End, function() {
        const model = codeEditor.getModel();
        const lastLine = model.getLineCount();
        const lastColumn = model.getLineMaxColumn(lastLine);
        codeEditor.setPosition({ lineNumber: lastLine, column: lastColumn });
    });
    
    // Ctrl+Right: 按单词向右跳转 (Monaco默认支持)
    // Ctrl+Left: 按单词向左跳转 (Monaco默认支持)
    
    // Ctrl+L: 选择整行
    codeEditor.addCommand(monaco.KeyMod.CtrlCmd | monaco.KeyCode.KeyL, function() {
        codeEditor.getAction('editor.action.selectLine').run();
    });
}

// 保持编辑器简洁，不添加复杂命令
function addCustomCommands() {
    // 编辑器保持朴素，不添加额外命令
    return;
}

// 更新编辑器语言和模板
function updateEditorLanguageAndTemplate(language) {
    if (!codeEditor) return;
    
    let monacoLang = 'go';
    let template = '';
    
    // 如果是核心模式，使用题目特定的模板
    if (currentProblem && currentProblem.mode === 'core' && codeTemplates) {
        switch(language) {
            case 'python':
                monacoLang = 'python';
                template = codeTemplates.python || '';
                break;
            case 'cpp':
                monacoLang = 'cpp';
                template = codeTemplates.cpp || '';
                break;
            case 'java':
                monacoLang = 'java';
                template = codeTemplates.java || '';
                break;
            default:
                monacoLang = 'go';
                template = codeTemplates.go || '';
        }
    } else {
        // 传统模式模板
        switch(language) {
            case 'python':
                monacoLang = 'python';
                template = 'def solution():\n    # Your code here\n    pass\n\nsolution()';
                break;
            case 'cpp':
                monacoLang = 'cpp';
                template = '#include <iostream>\n#include <vector>\nusing namespace std;\n\nint main() {\n    // Your code here\n    return 0;\n}';
                break;
            case 'java':
                monacoLang = 'java';
                template = 'public class Solution {\n    public static void main(String[] args) {\n        // Your code here\n    }\n}';
                break;
            default:
                monacoLang = 'go';
                template = 'package main\n\nimport "fmt"\n\nfunc main() {\n    // Your code here\n    m := map[int]int{}\n    fmt.Println(m)\n}';
        }
    }
    
    monaco.editor.setModelLanguage(codeEditor.getModel(), monacoLang);
    codeEditor.setValue(template);
    codeEditor.focus();
}

// 显示通知
function showNotification(message, type = 'success') {
    console.log('Notification:', message, type);
    const notification = document.getElementById('notification');
    const notificationText = document.getElementById('notificationText');
    
    if (!notification || !notificationText) {
        console.error('Notification elements not found');
        alert(message); // 备用方案
        return;
    }
    
    notificationText.textContent = message;
    
    // 设置颜色
    const notificationDiv = notification.querySelector('div');
    notificationDiv.className = `px-6 py-3 rounded-lg shadow-lg text-white`;
    
    switch(type) {
        case 'error':
            notificationDiv.classList.add('bg-red-500');
            break;
        case 'warning':
            notificationDiv.classList.add('bg-yellow-500');
            break;
        default:
            notificationDiv.classList.add('bg-green-500');
    }
    
    notification.classList.remove('hidden');
    
    setTimeout(() => {
        notification.classList.add('hidden');
    }, 3000);
}

// 显示登录表单
function showLogin() {
    console.log('Showing login form');
    hideAllSections();
    const loginForm = document.getElementById('loginForm');
    const authButtons = document.getElementById('authButtons');
    
    if (loginForm) {
        loginForm.classList.remove('hidden');
    } else {
        console.error('Login form not found');
    }
    
    if (authButtons) {
        authButtons.classList.remove('hidden');
    }
}

// 显示注册表单
function showRegister() {
    console.log('Showing register form');
    hideAllSections();
    const registerForm = document.getElementById('registerForm');
    const authButtons = document.getElementById('authButtons');
    
    if (registerForm) {
        registerForm.classList.remove('hidden');
    } else {
        console.error('Register form not found');
    }
    
    if (authButtons) {
        authButtons.classList.remove('hidden');
    }
}

// 显示主要内容
function showMainContent() {
    console.log('Showing main content');
    hideAllSections();
    const mainContent = document.getElementById('mainContent');
    const userInfo = document.getElementById('userInfo');
    const authButtons = document.getElementById('authButtons');
    
    if (mainContent) {
        mainContent.classList.remove('hidden');
    }
    if (userInfo) {
        userInfo.classList.remove('hidden');
    }
    if (authButtons) {
        authButtons.classList.add('hidden');
    }
    
    // 更新用户信息显示
    updateUserInfo();
    
    // 默认显示控制台
    showTab('dashboard');
}

// 隐藏所有sections
function hideAllSections() {
    const sections = ['loginForm', 'registerForm', 'mainContent', 'userInfo'];
    sections.forEach(sectionId => {
        const element = document.getElementById(sectionId);
        if (element) {
            element.classList.add('hidden');
        }
    });
}

// 用户注册
async function register(event) {
    console.log('Register function called');
    event.preventDefault();
    
    const username = document.getElementById('regUsername').value;
    const email = document.getElementById('regEmail').value;
    const password = document.getElementById('regPassword').value;
    
    console.log('Registration data:', { username, email });
    
    if (!username || !email || !password) {
        showNotification('请填写所有字段', 'error');
        return;
    }
    
    try {
        showNotification('正在注册...', 'warning');
        
        const response = await fetch(`${API_BASE}/users/register`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, email, password })
        });
        
        console.log('Register response status:', response.status);
        
        if (response.ok) {
            const data = await response.json();
            console.log('Registration successful:', data);
            showNotification('注册成功！请登录');
            showLogin();
        } else {
            const error = await response.json();
            console.error('Registration error:', error);
            showNotification(error.error || '注册失败', 'error');
        }
    } catch (error) {
        console.error('Network error during registration:', error);
        showNotification('网络错误，请重试', 'error');
    }
}

// 用户登录
async function login(event) {
    console.log('Login function called');
    event.preventDefault();
    
    const username = document.getElementById('loginUsername').value;
    const password = document.getElementById('loginPassword').value;
    
    console.log('Login data:', { username });
    
    if (!username || !password) {
        showNotification('请填写用户名和密码', 'error');
        return;
    }
    
    try {
        showNotification('正在登录...', 'warning');
        
        const response = await fetch(`${API_BASE}/users/login`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ username, password })
        });
        
        console.log('Login response status:', response.status);
        
        if (response.ok) {
            const data = await response.json();
            console.log('Login successful:', data);
            currentUser = data.user;
            localStorage.setItem('currentUser', JSON.stringify(currentUser));
            
            showNotification('登录成功！');
            showMainContent();
        } else {
            const error = await response.json();
            console.error('Login error:', error);
            showNotification(error.error || '登录失败', 'error');
        }
    } catch (error) {
        console.error('Network error during login:', error);
        showNotification('网络错误，请重试', 'error');
    }
}

// 用户退出
function logout() {
    currentUser = null;
    localStorage.removeItem('currentUser');
    
    // 停止匹配轮询
    stopMatchPolling();
    
    // 停止WebSocket连接（使用统一的断开函数）
    disconnectWebSocket();
    
    showNotification('已退出登录');
    showLogin();
}

// 更新用户信息显示
function updateUserInfo() {
    if (currentUser) {
        document.getElementById('username').textContent = currentUser.username;
        document.getElementById('rating').textContent = `${currentUser.rating} 分`;
        
        // 更新控制台信息
        document.getElementById('dashboardUsername').textContent = currentUser.username;
        document.getElementById('dashboardRating').textContent = `积分: ${currentUser.rating}`;
        document.getElementById('dashboardWins').textContent = `胜利: ${currentUser.wins}`;
        document.getElementById('dashboardLosses').textContent = `失败: ${currentUser.losses}`;
    }
}

// 刷新用户信息（从服务器重新获取最新数据）
async function refreshUserInfo() {
    if (!currentUser || !currentUser.id) {
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE}/users/profile/${currentUser.id}`);
        if (response.ok) {
            const data = await response.json();
            
            // 更新本地存储的用户信息
            currentUser = data.user;
            localStorage.setItem('currentUser', JSON.stringify(currentUser));
            
            // 更新界面显示
            updateUserInfo();
            
            console.log('User info refreshed:', currentUser);
        } else {
            console.error('Failed to refresh user info');
        }
    } catch (error) {
        console.error('Error refreshing user info:', error);
    }
}

// 切换选项卡
function showTab(tabName) {
    // 隐藏所有选项卡内容
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.classList.add('hidden');
    });
    
    // 显示选中的选项卡
    document.getElementById(tabName + 'Tab').classList.remove('hidden');
    
    // 根据选项卡加载相应数据
    switch(tabName) {
        case 'problems':
            loadProblems();
            break;
        case 'leaderboard':
            loadLeaderboard();
            break;
        case 'battle':
            resetBattleInterface();
            break;
    }
}

// 加载题目列表
async function loadProblems() {
    try {
        const response = await fetch(`${API_BASE}/problems`);
        if (response.ok) {
            const data = await response.json();
            displayProblems(data.problems);
        }
    } catch (error) {
        showNotification('加载题目失败', 'error');
    }
}

// 显示题目列表
function displayProblems(problems) {
    const container = document.getElementById('problemsList');
    container.innerHTML = '';
    
    problems.forEach((problem, index) => {
        const difficultyColor = {
            'easy': 'text-green-500',
            'medium': 'text-yellow-500',
            'hard': 'text-red-500'
        };
        
        const problemElement = document.createElement('div');
        problemElement.className = 'border-b pb-4 mb-4 last:border-b-0';
        problemElement.innerHTML = `
            <div class="flex justify-between items-start">
                <div class="flex-1">
                    <h3 class="text-lg font-semibold mb-2">${problem.title}</h3>
                    <p class="text-gray-600 mb-2">${problem.description.substring(0, 100)}...</p>
                    <div class="flex items-center space-x-4">
                        <span class="text-sm ${difficultyColor[problem.difficulty]} font-medium">
                            ${problem.difficulty.toUpperCase()}
                        </span>
                        <span class="text-sm text-gray-500">时间限制: ${problem.time_limit}ms</span>
                        <span class="text-sm text-gray-500">内存限制: ${problem.memory_limit}MB</span>
                    </div>
                </div>
                <button onclick="viewProblem(${problem.id})" class="bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded ml-4">
                    查看详情
                </button>
            </div>
        `;
        container.appendChild(problemElement);
    });
}

// 查看题目详情
async function viewProblem(problemId) {
    try {
        const response = await fetch(`${API_BASE}/problems/${problemId}`);
        if (response.ok) {
            const data = await response.json();
            // 这里可以显示题目详情模态框
            showNotification('功能开发中...', 'warning');
        }
    } catch (error) {
        showNotification('加载题目详情失败', 'error');
    }
}

// 加载排行榜
async function loadLeaderboard() {
    try {
        const response = await fetch(`${API_BASE}/users/leaderboard`);
        if (response.ok) {
            const data = await response.json();
            displayLeaderboard(data.leaderboard);
            
            // 排行榜加载完成后，刷新用户信息确保数据一致
            refreshUserInfo();
        }
    } catch (error) {
        showNotification('加载排行榜失败', 'error');
    }
}

// 显示排行榜
function displayLeaderboard(users) {
    const container = document.getElementById('leaderboardList');
    container.innerHTML = '';
    
    users.forEach((user, index) => {
        const rankIcon = index < 3 ? 
            ['🥇', '🥈', '🥉'][index] : 
            `#${index + 1}`;
            
        const userElement = document.createElement('div');
        userElement.className = 'flex items-center justify-between p-4 border-b last:border-b-0';
        userElement.innerHTML = `
            <div class="flex items-center space-x-4">
                <span class="text-2xl">${rankIcon}</span>
                <div>
                    <h3 class="font-semibold">${user.username}</h3>
                    <p class="text-sm text-gray-500">胜率: ${user.wins + user.losses > 0 ? Math.round(user.wins / (user.wins + user.losses) * 100) : 0}%</p>
                </div>
            </div>
            <div class="text-right">
                <p class="text-xl font-bold text-yellow-600">${user.rating}</p>
                <p class="text-sm text-gray-500">${user.wins}胜 ${user.losses}负</p>
            </div>
        `;
        container.appendChild(userElement);
    });
}

// 重置对战界面
function resetBattleInterface() {
    document.getElementById('matchmaking').classList.remove('hidden');
    document.getElementById('battleRoom').classList.add('hidden');
    document.getElementById('matchStatus').classList.add('hidden');
    document.getElementById('matchBtn').style.display = 'inline-block';
    document.getElementById('cancelBtn').classList.add('hidden');
    
    if (battleTimer) {
        clearInterval(battleTimer);
        battleTimer = null;
    }
    
    // 停止匹配轮询
    stopMatchPolling();
    
    // 断开WebSocket连接
    disconnectWebSocket();
    
    // 重置状态
    battleId = null;
    isInBattle = false;
}

// 开始匹配
async function startMatch() {
    if (!currentUser) {
        showNotification('请先登录', 'error');
        return;
    }
    
    document.getElementById('matchBtn').style.display = 'none';
    document.getElementById('cancelBtn').classList.remove('hidden');
    document.getElementById('matchStatus').classList.remove('hidden');
    
    // 建立WebSocket连接用于实时匹配和对战通知
    if (!ws || ws.readyState !== WebSocket.OPEN) {
        connectWebSocket();
        // 等待连接建立
        await new Promise((resolve) => {
            const checkConnection = () => {
                if (ws && ws.readyState === WebSocket.OPEN) {
                    resolve();
                } else {
                    setTimeout(checkConnection, 100);
                }
            };
            checkConnection();
        });
    }
    
    try {
        const response = await fetch(`${API_BASE}/match/join`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ user_id: currentUser.id })
        });
        
        if (response.ok) {
            // 设置匹配状态并开始轮询
            isMatching = true;
            pollMatchStatus();
        } else {
            showNotification('匹配失败', 'error');
            stopMatchPolling();
            resetBattleInterface();
        }
    } catch (error) {
        showNotification('网络错误', 'error');
        stopMatchPolling();
        resetBattleInterface();
    }
}

// 取消匹配
async function cancelMatch() {
    // 立即停止轮询，防止竞态条件
    stopMatchPolling();
    
    try {
        const response = await fetch(`${API_BASE}/match/leave`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ user_id: currentUser.id })
        });
        
        if (response.ok) {
            showNotification('已取消匹配', 'success');
        } else {
            showNotification('取消匹配失败', 'error');
        }
    } catch (error) {
        showNotification('网络错误', 'error');
    }
    
    // 重置界面和断开连接
    resetBattleInterface();
}

// 轮询匹配状态
async function pollMatchStatus() {
    // 如果不在匹配状态，停止轮询
    if (!isMatching) {
        console.log('Stopped polling match status - not in matching state');
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/match/status?user_id=${currentUser.id}`);
        if (response.ok) {
            const data = await response.json();
            
            if (data.status === 'waiting') {
                // 继续等待，但只有在仍然匹配状态时才继续
                if (isMatching) {
                    pollMatchStatusTimer = setTimeout(pollMatchStatus, 1000);
                }
            } else if (data.status.length > 10) {
                // 匹配成功，获得battleId
                battleId = data.status;
                isMatching = false; // 停止匹配状态
                showNotification('匹配成功！进入对战', 'success');
                enterBattleRoom();
            } else {
                // 状态为 idle 或其他，说明匹配已取消
                isMatching = false;
                console.log('Match cancelled by server, stopping poll');
            }
        }
    } catch (error) {
        console.error('Error polling match status:', error);
        // 只有在仍然匹配状态时才重试
        if (isMatching) {
            pollMatchStatusTimer = setTimeout(pollMatchStatus, 2000);
        }
    }
}

// 停止匹配状态轮询
function stopMatchPolling() {
    isMatching = false;
    if (pollMatchStatusTimer) {
        clearTimeout(pollMatchStatusTimer);
        pollMatchStatusTimer = null;
    }
    console.log('Match polling stopped');
}

// 进入对战房间
async function enterBattleRoom() {
    console.log('Entering battle room with battleId:', battleId);
    
    if (!battleId) {
        showNotification('战斗ID无效', 'error');
        return;
    }
    
    // 首先加入战斗房间
    try {
        const response = await fetch(`${API_BASE}/battle/${battleId}/join`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ user_id: currentUser.id })
        });
        
        if (!response.ok) {
            const error = await response.json();
            showNotification(error.error || '加入战斗失败', 'error');
            return;
        }
        
        console.log('Successfully joined battle room');
    } catch (error) {
        console.error('Error joining battle:', error);
        showNotification('加入战斗失败', 'error');
        return;
    }
    
    document.getElementById('matchmaking').classList.add('hidden');
    document.getElementById('battleRoom').classList.remove('hidden');
    
    // 加载对战信息
    await loadBattleInfo();
    
    // 开始计时器
    startBattleTimer();
}

// 加载对战信息
async function loadBattleInfo() {
    try {
        const response = await fetch(`${API_BASE}/battle/${battleId}`);
        if (response.ok) {
            const data = await response.json();
            // 更新界面信息
            updateBattleDisplay(data.battle);
            
            // 加载战斗中的特定题目
            if (data.battle && data.battle.problem_id) {
                await loadBattleProblem(data.battle.problem_id);
            } else {
                // 如果没有题目，加载随机题目
                await loadRandomProblem();
            }
        }
    } catch (error) {
        console.error('加载对战信息失败:', error);
        // 如果加载失败，回退到随机题目
        await loadRandomProblem();
    }
}

// 加载战斗中的特定题目
async function loadBattleProblem(problemId) {
    try {
        const response = await fetch(`${API_BASE}/problems/${problemId}`);
        if (response.ok) {
            const data = await response.json();
            
            // 显示题目
            displayProblem(data.problem);
            
            // 如果是核心模式，加载额外数据
            if (data.problem.mode === 'core') {
                codeTemplates = data.code_templates;
                functionSignature = data.function_signature;
                coreTestCases = data.core_test_cases;
                
                console.log('Loaded core mode problem:', data.problem.title);
                console.log('Function signature:', functionSignature);
                
                // 更新编辑器模板
                const languageSelect = document.getElementById('languageSelect');
                if (languageSelect && codeTemplates) {
                    updateEditorLanguageAndTemplate(languageSelect.value);
                }
            } else {
                // 清空核心模式数据
                codeTemplates = null;
                functionSignature = null;
                coreTestCases = null;
                console.log('Loaded traditional problem:', data.problem.title);
            }
        } else {
            console.error('Failed to load battle problem:', problemId);
            await loadRandomProblem();
        }
    } catch (error) {
        console.error('Error loading battle problem:', error);
        await loadRandomProblem();
    }
}

// 加载随机题目
async function loadRandomProblem() {
    try {
        const response = await fetch(`${API_BASE}/problems`);
        if (response.ok) {
            const data = await response.json();
            if (data.problems && data.problems.length > 0) {
                const randomProblem = data.problems[Math.floor(Math.random() * data.problems.length)];
                displayProblem(randomProblem);
            }
        }
    } catch (error) {
        console.error('加载题目失败:', error);
    }
}

// 显示题目
function displayProblem(problem) {
    document.getElementById('problemTitle').textContent = problem.title;
    document.getElementById('problemDescription').textContent = problem.description;
    
    // 保存当前题目信息
    currentProblem = problem;
    
    // 根据题目模式显示不同的示例
    if (problem.mode === 'core') {
        // 核心模式：显示函数调用示例
        try {
            const testCases = JSON.parse(problem.test_cases);
            if (testCases && testCases.length > 0) {
                const example = testCases[0];
                let exampleText = '';
                
                // 构建示例文本
                if (example.args && example.expected !== undefined) {
                    const argsStr = example.args.map(arg => 
                        Array.isArray(arg) ? JSON.stringify(arg) : arg
                    ).join(', ');
                    
                    const expectedStr = Array.isArray(example.expected) ? 
                        JSON.stringify(example.expected) : example.expected;
                    
                    exampleText = `调用: ${problem.function_signature?.function_name || 'function'}(${argsStr})\n输出: ${expectedStr}`;
                    
                    if (example.explain) {
                        exampleText += `\n说明: ${example.explain}`;
                    }
                }
                
                document.getElementById('problemExample').textContent = exampleText;
            }
        } catch (error) {
            console.error('解析核心模式测试用例失败:', error);
        }
    } else {
        // 传统模式：显示输入输出示例
        try {
            const testCases = JSON.parse(problem.test_cases);
            if (testCases && testCases.length > 0) {
                const example = testCases[0];
                document.getElementById('problemExample').textContent = 
                    `输入: ${example.input}\n输出: ${example.output}`;
            }
        } catch (error) {
            console.error('解析测试用例失败:', error);
        }
    }
}

// 更新对战显示
function updateBattleDisplay(battle) {
    if (!battle || !battle.player1 || !battle.player2) {
        console.error('Invalid battle data:', battle);
        return;
    }
    
    // 确定当前用户是player1还是player2
    let currentPlayer, opponent;
    if (battle.player1.id === currentUser.id) {
        currentPlayer = battle.player1;
        opponent = battle.player2;
    } else {
        currentPlayer = battle.player2;
        opponent = battle.player1;
    }
    
    // 更新当前用户信息
    document.getElementById('player1Name').textContent = currentPlayer.username;
    document.getElementById('player1Rating').textContent = `Rating: ${currentPlayer.rating}`;
    
    // 更新对手信息
    document.getElementById('player2Name').textContent = opponent.username;
    document.getElementById('player2Rating').textContent = `Rating: ${opponent.rating}`;
    
    // 显示对战状态
    if (battle.status === 'active') {
        showNotification(`对战开始！对手是 ${opponent.username}`, 'success');
    }
    
    console.log('Battle display updated:', {
        current: currentPlayer,
        opponent: opponent,
        status: battle.status
    });
}

// 开始对战计时器
function startBattleTimer() {
    startTime = Date.now();
    battleTimer = setInterval(() => {
        const elapsed = Math.floor((Date.now() - startTime) / 1000);
        const minutes = Math.floor(elapsed / 60);
        const seconds = elapsed % 60;
        document.getElementById('battleTimer').textContent = 
            `${minutes.toString().padStart(2, '0')}:${seconds.toString().padStart(2, '0')}`;
    }, 1000);
}

// 连接WebSocket
// WebSocket相关变量
let reconnectAttempts = 0;
let maxReconnectAttempts = 5;
let reconnectDelay = 3000;
let heartbeatInterval = null;
let isManualClose = false;
let shouldMaintainConnection = false; // 是否应该保持WebSocket连接

// 匹配状态相关变量
let pollMatchStatusTimer = null;
let isMatching = false;

function connectWebSocket() {
    if (ws) {
        isManualClose = true; // 标记为手动关闭，避免重连
        ws.close();
        isManualClose = false;
    }
    
    if (!currentUser || !currentUser.id) {
        console.error('Cannot connect WebSocket: user not logged in');
        return;
    }
    
    // 设置应该保持连接的状态
    shouldMaintainConnection = true;
    
    ws = new WebSocket(`ws://localhost:8080/ws?user_id=${currentUser.id}`);
    
    ws.onopen = function() {
        console.log('WebSocket连接已建立');
        reconnectAttempts = 0; // 重置重连次数
        
        // 启动心跳检测
        startHeartbeat();
    };
    
    ws.onmessage = function(event) {
        const data = JSON.parse(event.data);
        
        // 处理心跳响应
        if (data.type === 'pong') {
            return; // 心跳响应，不需要额外处理
        }
        
        console.log('Received WebSocket message:', data);
        handleWebSocketMessage(data);
    };
    
    ws.onclose = function(event) {
        console.log('WebSocket连接已关闭', event.code, event.reason);
        
        // 停止心跳
        stopHeartbeat();
        
        // 只有在应该保持连接且不是手动关闭时才重连
        if (!isManualClose && shouldMaintainConnection && currentUser && currentUser.id && reconnectAttempts < maxReconnectAttempts) {
            reconnectAttempts++;
            console.log(`尝试重连 (${reconnectAttempts}/${maxReconnectAttempts})...`);
            
            setTimeout(() => {
                if (shouldMaintainConnection && currentUser && currentUser.id) {
                    connectWebSocket();
                }
            }, reconnectDelay * reconnectAttempts); // 递增延迟
        } else if (reconnectAttempts >= maxReconnectAttempts) {
            console.warn('WebSocket重连次数已达上限，停止重连');
            showNotification('网络连接不稳定，请刷新页面重试', 'error');
        } else if (!shouldMaintainConnection) {
            console.log('WebSocket connection not needed, no reconnection');
        }
    };
    
    ws.onerror = function(error) {
        console.error('WebSocket错误:', error);
    };
}

// 启动心跳检测
function startHeartbeat() {
    stopHeartbeat(); // 先停止之前的心跳
    
    heartbeatInterval = setInterval(() => {
        if (ws && ws.readyState === WebSocket.OPEN) {
            ws.send(JSON.stringify({ type: 'ping' }));
        }
    }, 30000); // 每30秒发送一次心跳
}

// 停止心跳检测
function stopHeartbeat() {
    if (heartbeatInterval) {
        clearInterval(heartbeatInterval);
        heartbeatInterval = null;
    }
}

// 主动断开WebSocket连接
function disconnectWebSocket() {
    // 设置不应该保持连接
    shouldMaintainConnection = false;
    
    if (ws) {
        isManualClose = true;
        stopHeartbeat();
        ws.close();
        ws = null;
        isManualClose = false;
        reconnectAttempts = 0;
        console.log('WebSocket manually disconnected');
    }
}

// 处理WebSocket消息
function handleWebSocketMessage(data) {
    switch(data.type) {
        case 'battle_update':
            updateBattleDisplay(data.battle);
            break;
        case 'opponent_submitted':
        case 'code_submitted':
            // 检查是否是对手提交的代码
            if (data.data && data.data.user_id !== currentUser.id) {
                showNotification('对手已提交代码！', 'warning');
            }
            break;
        case 'battle_finished':
            showNotification(`对战结束！${data.winner ? '恭喜获胜！' : '很遗憾失败了'}`, 
                data.winner ? 'success' : 'error');
            break;
        case 'battle_end':
            handleBattleEnd(data.data);
            break;
        case 'judge_result':
            handleJudgeResult(data.data);
            break;
        case 'battle_flee':
            handleBattleFlee(data.data);
            break;
        case 'opponent_fled':
            handleOpponentFled(data.data);
            break;
    }
}

// 处理对战结束
function handleBattleEnd(battleData) {
    console.log('Battle ended:', battleData);
    
    const winner = battleData.winner;
    const loser = battleData.loser;
    const battleDuration = battleData.battle_duration;
    
    // 构建对战结果显示
    let resultHTML = `
        <div class="bg-white p-6 rounded-lg shadow-lg max-w-md mx-auto mt-8">
            <div class="text-center">
                <h2 class="text-2xl font-bold mb-4">对战结束</h2>
    `;
    
    if (winner) {
        const isWinner = winner.id === currentUser.id;
        resultHTML += `
            <div class="mb-4">
                <div class="text-lg ${isWinner ? 'text-green-600' : 'text-red-600'} font-semibold">
                    ${isWinner ? '🏆 恭喜获胜！' : '💔 很遗憾失败了'}
                </div>
                <div class="mt-2 text-gray-600">
                    获胜者: ${winner.username} (${winner.rating}分)
                </div>
                <div class="text-gray-600">
                    失败者: ${loser.username} (${loser.rating}分)
                </div>
            </div>
        `;
    } else {
        resultHTML += `
            <div class="mb-4">
                <div class="text-lg text-gray-600 font-semibold">
                    📝 平局
                </div>
                <div class="mt-2 text-gray-600">
                    双方都未能完成题目
                </div>
            </div>
        `;
    }
    
    if (battleDuration) {
        resultHTML += `
            <div class="mb-4 text-gray-600">
                对战时长: ${Math.floor(battleDuration / 60)}分${Math.floor(battleDuration % 60)}秒
            </div>
        `;
    }
    
    // 显示提交结果
    if (battleData.player1_result || battleData.player2_result) {
        resultHTML += `
            <div class="mb-4 text-sm">
                <div class="border-t pt-3">
                    <h3 class="font-semibold mb-2">提交结果:</h3>
                    <div class="space-y-1">
                        <div>Player 1: ${battleData.player1_result || '未提交'}</div>
                        <div>Player 2: ${battleData.player2_result || '未提交'}</div>
                    </div>
                </div>
            </div>
        `;
    }
    
    resultHTML += `
                <div class="flex space-x-3">
                    <button 
                        onclick="startMatch()" 
                        class="flex-1 bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">
                        再来一局
                    </button>
                    <button 
                        onclick="resetBattleInterface()" 
                        class="flex-1 bg-gray-500 hover:bg-gray-600 text-white px-4 py-2 rounded">
                        返回
                    </button>
                </div>
            </div>
        </div>
    `;
    
    // 显示结果对话框
    const resultContainer = document.createElement('div');
    resultContainer.className = 'fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';
    resultContainer.innerHTML = resultHTML;
    
    document.body.appendChild(resultContainer);
    
    // 点击背景关闭对话框
    resultContainer.addEventListener('click', (e) => {
        if (e.target === resultContainer) {
            document.body.removeChild(resultContainer);
        }
    });
    
    // 显示通知
    if (winner && winner.id === currentUser.id) {
        showNotification('🎉 恭喜获胜！积分已增加', 'success');
    } else if (winner) {
        showNotification('💔 很遗憾失败了，继续努力！', 'error');
    } else {
        showNotification('📝 对战平局', 'warning');
    }
    
    // 重置对战状态
    battleId = null;
    isInBattle = false;
    
    // 重新获取用户信息以更新分数
    refreshUserInfo();
    
    // 对战结束后立即断开WebSocket连接
    disconnectWebSocket();
}

// 处理判题结果
function handleJudgeResult(result) {
    const resultDiv = document.getElementById('codeResult');
    if (!resultDiv) return;
    
    let statusClass = '';
    let statusIcon = '';
    
    switch(result.status) {
        case 'Accepted':
        case 'AC':
            statusClass = 'bg-green-100 border-green-400 text-green-700';
            statusIcon = 'fa-check-circle';
            break;
        case 'Wrong Answer':
        case 'WA':
            statusClass = 'bg-red-100 border-red-400 text-red-700';
            statusIcon = 'fa-times-circle';
            break;
        case 'Time Limit Exceeded':
        case 'TLE':
            statusClass = 'bg-yellow-100 border-yellow-400 text-yellow-700';
            statusIcon = 'fa-clock';
            break;
        case 'Memory Limit Exceeded':
        case 'MLE':
            statusClass = 'bg-purple-100 border-purple-400 text-purple-700';
            statusIcon = 'fa-memory';
            break;
        case 'Runtime Error':
        case 'RE':
            statusClass = 'bg-orange-100 border-orange-400 text-orange-700';
            statusIcon = 'fa-exclamation-triangle';
            break;
        case 'Compile Error':
        case 'CE':
            statusClass = 'bg-gray-100 border-gray-400 text-gray-700';
            statusIcon = 'fa-code';
            break;
        default:
            statusClass = 'bg-blue-100 border-blue-400 text-blue-700';
            statusIcon = 'fa-info-circle';
    }
    
    let resultHTML = `
        <div class="${statusClass} border px-4 py-3 rounded">
            <div class="flex items-center mb-2">
                <i class="fas ${statusIcon} mr-2"></i>
                <strong>${result.message || result.status}</strong>
            </div>
    `;
    
    if (result.execution_time) {
        resultHTML += `<div class="text-sm">执行时间: ${result.execution_time}ms</div>`;
    }
    
    if (result.memory_usage) {
        resultHTML += `<div class="text-sm">内存使用: ${result.memory_usage}KB</div>`;
    }
    
    if (result.error_msg) {
        // 如果是WA状态且包含diff信息，特殊处理
        if (result.status === 'WA' && result.error_msg.includes('Expected vs Actual')) {
            resultHTML += `<div class="text-sm mt-2">
                <strong>详细差异分析:</strong>
                <pre class="bg-gray-50 p-3 rounded mt-1 text-xs font-mono whitespace-pre-wrap overflow-x-auto">${result.error_msg}</pre>
            </div>`;
        } else {
            resultHTML += `<div class="text-sm mt-2 font-mono bg-gray-50 p-2 rounded">${result.error_msg}</div>`;
        }
    }
    
    if (result.output && result.status !== 'AC') {
        resultHTML += `<div class="text-sm mt-2">
            <strong>输出:</strong>
            <pre class="bg-gray-50 p-2 rounded mt-1 text-xs">${result.output}</pre>
        </div>`;
    }
    
    resultHTML += '</div>';
    
    resultDiv.innerHTML = resultHTML;
    resultDiv.classList.remove('hidden');
    
    // 显示通知
    const notification = result.status === 'AC' ? 'success' : 'error';
    showNotification(result.message || result.status, notification);
}

// 检查语法
async function checkSyntax() {
    let code = '';
    if (codeEditor && codeEditor.getValue) {
        code = codeEditor.getValue();
    } else {
        // 回退到textarea
        const fallbackEditor = document.getElementById('fallbackEditor');
        code = fallbackEditor ? fallbackEditor.value : '';
    }
    const language = document.getElementById('languageSelect').value;
    
    if (!code.trim()) {
        showNotification('请输入代码', 'warning');
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE}/judge/check`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ code, language })
        });
        
        if (response.ok) {
            const data = await response.json();
            showNotification('语法检查通过！', 'success');
        } else {
            const error = await response.json();
            showNotification(error.error || '语法检查失败', 'error');
        }
    } catch (error) {
        showNotification('网络错误', 'error');
    }
}

// 轮询判题结果（WebSocket备用方案）
async function pollForJudgeResult() {
    if (!battleId) return;
    
    try {
        const response = await fetch(`${API_BASE}/battle/${battleId}`);
        if (response.ok) {
            const data = await response.json();
            const battle = data.battle;
            
            // 检查当前用户的判题结果
            if (currentUser && battle.results && battle.results[currentUser.id]) {
                const result = battle.results[currentUser.id];
                if (result.status && result.status !== 'Pending') {
                    // 判题完成，显示结果
                    handleJudgeResult({
                        user_id: currentUser.id,
                        status: result.status,
                        execution_time: result.execution_time,
                        message: getStatusMessage(result.status)
                    });
                    return;
                }
            }
            
            // 判题还在进行中，继续轮询
            setTimeout(() => pollForJudgeResult(), 1000);
        }
    } catch (error) {
        console.error('Polling error:', error);
        // 最多轮询30秒
        if (Date.now() - pollStartTime < 30000) {
            setTimeout(() => pollForJudgeResult(), 2000);
        } else {
            const resultDiv = document.getElementById('codeResult');
            if (resultDiv) {
                resultDiv.innerHTML = `
                    <div class="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded">
                        <i class="fas fa-exclamation-triangle mr-2"></i>
                        判题超时，请稍后查看结果
                    </div>
                `;
            }
        }
    }
}

// 获取状态消息的辅助函数
function getStatusMessage(status) {
    switch(status) {
        case 'Accepted':
        case 'AC':
            return '恭喜！代码通过所有测试用例';
        case 'Wrong Answer':
        case 'WA':
            return '答案错误，请检查逻辑';
        case 'Time Limit Exceeded':
        case 'TLE':
            return '超出时间限制';
        case 'Memory Limit Exceeded':
        case 'MLE':
            return '超出内存限制';
        case 'Runtime Error':
        case 'RE':
            return '运行时错误';
        case 'Compile Error':
        case 'CE':
            return '编译错误';
        default:
            return '判题完成';
    }
}

// 提交代码
async function submitCode() {
    let code = '';
    if (codeEditor && codeEditor.getValue) {
        code = codeEditor.getValue();
    } else {
        // 回退到textarea
        const fallbackEditor = document.getElementById('fallbackEditor');
        code = fallbackEditor ? fallbackEditor.value : '';
    }
    const language = document.getElementById('languageSelect').value;
    
    if (!code.trim()) {
        showNotification('请输入代码', 'warning');
        return;
    }
    
    if (!battleId) {
        showNotification('请先开始对战', 'error');
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE}/battle/${battleId}/submit`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ 
                user_id: currentUser.id,
                code, 
                language 
            })
        });
        
        if (response.ok) {
            showNotification('代码提交成功！等待判题结果...', 'success');
            
            // 显示结果
            const resultDiv = document.getElementById('codeResult');
            resultDiv.innerHTML = `
                <div class="bg-blue-100 border border-blue-400 text-blue-700 px-4 py-3 rounded">
                    <i class="fas fa-spinner fa-spin mr-2"></i>
                    正在判题中...
                </div>
            `;
            resultDiv.classList.remove('hidden');
            
            // 备用方案：如果WebSocket没有连接，使用轮询
            if (!ws || ws.readyState !== WebSocket.OPEN) {
                console.log('WebSocket not connected, using polling fallback');
                window.pollStartTime = Date.now();
                setTimeout(() => pollForJudgeResult(), 2000);
            }
            
        } else {
            const error = await response.json();
            showNotification(error.error || '提交失败', 'error');
        }
    } catch (error) {
        showNotification('网络错误', 'error');
    }
}

// 逃跑功能
async function fleeBattle() {
    if (!battleId) {
        showNotification('当前没有进行中的对战', 'error');
        return;
    }

    // 确认对话框
    if (!confirm('确定要逃跑吗？这将导致对战失败并扣除积分！')) {
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/battle/${battleId}/flee`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify({ 
                user_id: currentUser.id
            })
        });

        if (response.ok) {
            console.log('Successfully fled from battle');
            // WebSocket 会处理后续的通知和界面更新
        } else {
            const error = await response.json();
            showNotification(error.error || '逃跑失败', 'error');
        }
    } catch (error) {
        console.error('Error fleeing battle:', error);
        showNotification('网络错误，逃跑失败', 'error');
    }
}

// 处理自己逃跑的结果
function handleBattleFlee(data) {
    console.log('Handle battle flee:', data);
    
    showNotification(data.message, 'error');
    
    // 显示逃跑结果
    const resultHTML = `
        <div class="bg-white p-6 rounded-lg shadow-lg max-w-md mx-auto mt-8">
            <div class="text-center">
                <h2 class="text-2xl font-bold mb-4 text-red-600">对战逃跑</h2>
                <div class="mb-4">
                    <div class="text-lg text-red-600 font-semibold">
                        😞 很遗憾，你选择了逃跑
                    </div>
                    <div class="mt-2 text-gray-600">
                        ${data.message}
                    </div>
                </div>
                <div class="flex space-x-3">
                    <button 
                        onclick="startMatch()" 
                        class="flex-1 bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">
                        再来一局
                    </button>
                    <button 
                        onclick="resetBattleInterface()" 
                        class="flex-1 bg-gray-500 hover:bg-gray-600 text-white px-4 py-2 rounded">
                        返回
                    </button>
                </div>
            </div>
        </div>
    `;
    
    showBattleResult(resultHTML);
    
    // 重置对战状态并刷新用户信息
    battleId = null;
    isInBattle = false;
    refreshUserInfo();
    disconnectWebSocket();
}

// 处理对手逃跑的结果
function handleOpponentFled(data) {
    console.log('Handle opponent fled:', data);
    
    showNotification(data.message, 'success');
    
    // 显示获胜结果
    const resultHTML = `
        <div class="bg-white p-6 rounded-lg shadow-lg max-w-md mx-auto mt-8">
            <div class="text-center">
                <h2 class="text-2xl font-bold mb-4 text-green-600">不战而胜</h2>
                <div class="mb-4">
                    <div class="text-lg text-green-600 font-semibold">
                        🎉 恭喜！对手逃跑，你获得胜利
                    </div>
                    <div class="mt-2 text-gray-600">
                        ${data.message}
                    </div>
                    <div class="mt-2 text-gray-600">
                        对手: ${data.opponent ? data.opponent.username : '未知'}
                    </div>
                </div>
                <div class="flex space-x-3">
                    <button 
                        onclick="startMatch()" 
                        class="flex-1 bg-blue-500 hover:bg-blue-600 text-white px-4 py-2 rounded">
                        再来一局
                    </button>
                    <button 
                        onclick="resetBattleInterface()" 
                        class="flex-1 bg-gray-500 hover:bg-gray-600 text-white px-4 py-2 rounded">
                        返回
                    </button>
                </div>
            </div>
        </div>
    `;
    
    showBattleResult(resultHTML);
    
    // 重置对战状态并刷新用户信息
    battleId = null;
    isInBattle = false;
    refreshUserInfo();
    disconnectWebSocket();
}

// 显示对战结果对话框
function showBattleResult(resultHTML) {
    // 移除之前的结果对话框
    const existingResult = document.querySelector('.battle-result-modal');
    if (existingResult) {
        document.body.removeChild(existingResult);
    }
    
    const resultContainer = document.createElement('div');
    resultContainer.className = 'battle-result-modal fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50';
    resultContainer.innerHTML = resultHTML;
    
    document.body.appendChild(resultContainer);
    
    // 点击背景关闭对话框
    resultContainer.addEventListener('click', (e) => {
        if (e.target === resultContainer) {
            document.body.removeChild(resultContainer);
        }
    });
}

 