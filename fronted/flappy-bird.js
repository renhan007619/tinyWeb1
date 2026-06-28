// ==================== Flappy Bird 游戏逻辑 ====================
    // 游戏配置
    var CANVAS_WIDTH = 400;
    var CANVAS_HEIGHT = 600;
    var GRAVITY = 0.06;
    var JUMP_STRENGTH = -2.4;
    var PIPE_SPEED = 1.2;
    var PIPE_SPAWN_RATE = 180;
    var PIPE_GAP = 140;
    var PIPE_WIDTH = 60;
    
    // 游戏状态
    var canvas = document.getElementById('gameCanvas');
    var ctx = canvas.getContext('2d');
    
    // 兼容旧的DOM引用
    var gameScore = document.getElementById('gameScore') || document.getElementById('scoreDisplay');
    var startHint = document.getElementById('gameStartHint') || document.getElementById('gameStatus');
    var gameOverScreen = document.getElementById('gameOverScreen');
    
    // 游戏变量
    var bird = { x: 80, y: 250, velocity: 0, radius: 16, rotation: 0, wingAngle: 0 };
    var pipes = [];
    var score = 0;
    var frameCount = 0;
    var isPlaying = false;
    var isPaused = false;
    var animationId = null;
    var gameStartTime = null;
    var pipesPassed = 0;
    
    // 音效系统
    var audioCtx = null;
    var bgmOscillators = [];
    var bgmGains = []; // 存储gain节点以便清理
    var isMuted = false;
    
    // 初始化音频上下文
    function initAudio() {
        try {
            audioCtx = new (window.AudioContext || window.webkitAudioContext)();
        } catch (e) {
            console.log('音频初始化失败:', e);
            audioCtx = null;
        }
    }
    
    // 播放跳跃音效
    function playJumpSound() {
        if (!audioCtx || isMuted) return;
        try {
            var osc = audioCtx.createOscillator();
            var gain = audioCtx.createGain();
            osc.connect(gain);
            gain.connect(audioCtx.destination);
            
            osc.frequency.value = 400;
            osc.frequency.exponentialRampToValueAtTime(600, audioCtx.currentTime + 0.1);
            
            gain.gain.value = 0.3;
            gain.gain.exponentialRampToValueAtTime(0.01, audioCtx.currentTime + 0.1);
            
            osc.start(audioCtx.currentTime);
            osc.stop(audioCtx.currentTime + 0.1);
        } catch (e) {
            console.log('音效播放失败:', e);
        }
    }
    
    // 播放得分音效
    function playScoreSound() {
        if (!audioCtx || isMuted) return;
        try {
            var osc = audioCtx.createOscillator();
            var gain = audioCtx.createGain();
            osc.connect(gain);
            gain.connect(audioCtx.destination);
            
            osc.frequency.value = 800;
            
            gain.gain.value = 0.3;
            gain.gain.exponentialRampToValueAtTime(0.01, audioCtx.currentTime + 0.1);
            
            osc.start(audioCtx.currentTime);
            osc.stop(audioCtx.currentTime + 0.1);
        } catch (e) {
            console.log('音效播放失败:', e);
        }
    }
    
    // 播放游戏结束音效
    function playGameOverSound() {
        if (!audioCtx || isMuted) return;
        try {
            var osc = audioCtx.createOscillator();
            var gain = audioCtx.createGain();
            osc.connect(gain);
            gain.connect(audioCtx.destination);
            
            osc.frequency.value = 400;
            osc.frequency.exponentialRampToValueAtTime(100, audioCtx.currentTime + 0.5);
            
            gain.gain.value = 0.4;
            gain.gain.exponentialRampToValueAtTime(0.01, audioCtx.currentTime + 0.5);
            
            osc.start(audioCtx.currentTime);
            osc.stop(audioCtx.currentTime + 0.5);
        } catch (e) {
            console.log('音效播放失败:', e);
        }
    }
    
    // 背景音乐
    function startBGM() {
        if (!audioCtx || isMuted) return;
        // 先停止之前的背景音乐，防止重复
        stopBGM();
        try {
            // 简单的背景节奏 - 只创建一个低频嗡鸣，避免电流音
            var osc = audioCtx.createOscillator();
            var gain = audioCtx.createGain();
            osc.connect(gain);
            gain.connect(audioCtx.destination);
            
            // 使用三角波，更柔和
            osc.type = 'triangle';
            osc.frequency.value = 150;
            gain.gain.value = 0.03;
            
            osc.start(audioCtx.currentTime);
            bgmOscillators.push(osc);
            bgmGains.push(gain);
        } catch (e) {
            console.log('背景音乐启动失败:', e);
        }
    }
    
    function stopBGM() {
        // 停止振荡器
        bgmOscillators.forEach(function(osc) {
            try {
                osc.stop();
                osc.disconnect();
            } catch (e) {}
        });
        // 断开增益节点
        bgmGains.forEach(function(gain) {
            try {
                gain.disconnect();
            } catch (e) {}
        });
        bgmOscillators = [];
        bgmGains = [];
    }
    
    // 配色方案 - 使用琥珀橙金色调
    function getColors() {
        return {
            bg: ['#1a2332', '#2d3748'],
            pipe: '#2d3436',
            pipeCap: '#636e72',
            bird: '#e8a86d',
            birdEye: '#fff',
            birdBeak: '#f5a623',
            ground: '#2d3436',
            groundTop: '#636e72'
        };
    }
    
    // 绘制背景
    function drawBackground() {
        var colors = getColors();
        var gradient = ctx.createLinearGradient(0, 0, 0, CANVAS_HEIGHT);
        gradient.addColorStop(0, colors.bg[0]);
        gradient.addColorStop(1, colors.bg[1]);
        ctx.fillStyle = gradient;
        ctx.fillRect(0, 0, CANVAS_WIDTH, CANVAS_HEIGHT);
        
        // 绘制装饰星星
        ctx.fillStyle = 'rgba(232, 168, 109, 0.15)';
        for (var i = 0; i < 8; i++) {
            var x = (frameCount * 0.3 + i * 50) % (CANVAS_WIDTH + 50) - 25;
            var y = 50 + i * 35;
            ctx.beginPath();
            ctx.arc(x, y, 3, 0, Math.PI * 2);
            ctx.fill();
        }
    }
    
    // 绘制地面
    function drawGround() {
        var colors = getColors();
        var groundHeight = 80;
        
        // 地面主体
        ctx.fillStyle = colors.ground;
        ctx.fillRect(0, CANVAS_HEIGHT - groundHeight, CANVAS_WIDTH, groundHeight);
        
        // 地面顶部线条
        ctx.fillStyle = colors.groundTop;
        ctx.fillRect(0, CANVAS_HEIGHT - groundHeight, CANVAS_WIDTH, 4);
        
        // 地面纹理
        ctx.fillStyle = 'rgba(232, 168, 109, 0.2)';
        for (var i = 0; i < CANVAS_WIDTH; i += 20) {
            var offset = (frameCount * PIPE_SPEED) % 20;
            ctx.fillRect(i - offset, CANVAS_HEIGHT - groundHeight + 10, 10, 3);
            ctx.fillRect(i - offset + 5, CANVAS_HEIGHT - groundHeight + 25, 8, 2);
        }
    }
    
    // 绘制小鸟
    function drawBird() {
        var colors = getColors();
        
        ctx.save();
        ctx.translate(bird.x, bird.y);
        ctx.rotate(bird.rotation);
        
        // 身体
        ctx.fillStyle = colors.bird;
        ctx.beginPath();
        ctx.arc(0, 0, bird.radius, 0, Math.PI * 2);
        ctx.fill();
        
        // 身体高光
        ctx.fillStyle = 'rgba(255,255,255,0.3)';
        ctx.beginPath();
        ctx.arc(-4, -4, 6, 0, Math.PI * 2);
        ctx.fill();
        
        // 眼睛
        ctx.fillStyle = colors.birdEye;
        ctx.beginPath();
        ctx.arc(6, -4, 5, 0, Math.PI * 2);
        ctx.fill();
        
        // 眼珠
        ctx.fillStyle = '#1a1a2e';
        ctx.beginPath();
        ctx.arc(8, -4, 2.5, 0, Math.PI * 2);
        ctx.fill();
        
        // 眼珠高光
        ctx.fillStyle = '#fff';
        ctx.beginPath();
        ctx.arc(9, -5, 1, 0, Math.PI * 2);
        ctx.fill();
        
        // 嘴巴
        ctx.fillStyle = colors.birdBeak;
        ctx.beginPath();
        ctx.moveTo(10, 2);
        ctx.lineTo(18, 5);
        ctx.lineTo(10, 8);
        ctx.closePath();
        ctx.fill();
        
        // 翅膀
        ctx.save();
        ctx.translate(-6, 2);
        ctx.rotate(bird.wingAngle);
        ctx.fillStyle = '#d4956d';
        ctx.beginPath();
        ctx.ellipse(0, 0, 9, 6, 0, 0, Math.PI * 2);
        ctx.fill();
        ctx.fillStyle = 'rgba(0,0,0,0.1)';
        ctx.beginPath();
        ctx.ellipse(2, 2, 9, 6, 0, 0, Math.PI * 2);
        ctx.fill();
        ctx.restore();
        
        ctx.restore();
    }
    
    // 绘制管道
    function drawPipe(pipe) {
        var colors = getColors();
        
        var gradient = ctx.createLinearGradient(pipe.x, 0, pipe.x + PIPE_WIDTH, 0);
        gradient.addColorStop(0, colors.pipe);
        gradient.addColorStop(0.5, colors.pipeCap);
        gradient.addColorStop(1, colors.pipe);
        
        // 上管道
        ctx.fillStyle = gradient;
        ctx.fillRect(pipe.x, 0, PIPE_WIDTH, pipe.topHeight);
        
        // 上管道边缘
        ctx.fillStyle = colors.pipeCap;
        ctx.fillRect(pipe.x - 3, pipe.topHeight - 20, PIPE_WIDTH + 6, 20);
        
        // 下管道
        ctx.fillStyle = gradient;
        ctx.fillRect(pipe.x, pipe.topHeight + PIPE_GAP, PIPE_WIDTH, CANVAS_HEIGHT - pipe.topHeight - PIPE_GAP - 80);
        
        // 下管道边缘
        ctx.fillStyle = colors.pipeCap;
        ctx.fillRect(pipe.x - 3, pipe.topHeight + PIPE_GAP, PIPE_WIDTH + 6, 20);
        
        // 管道高光
        ctx.fillStyle = 'rgba(232, 168, 109, 0.15)';
        ctx.fillRect(pipe.x + 5, 0, 4, pipe.topHeight);
        ctx.fillRect(pipe.x + 5, pipe.topHeight + PIPE_GAP, 4, CANVAS_HEIGHT - pipe.topHeight - PIPE_GAP - 50);
    }
    
    // 更新游戏
    function update() {
        if (!isPlaying || isPaused) return;
        
        frameCount++;
        
        // 更新小鸟
        bird.velocity += GRAVITY;
        bird.y += bird.velocity;
        
        // 更新旋转角度
        bird.rotation = Math.min(Math.PI / 4, Math.max(-Math.PI / 4, bird.velocity * 0.08));
        
        // 翅膀扇动动画
        if (bird.velocity < 0) {
            bird.wingAngle = Math.sin(frameCount * 0.4) * 0.6;
        } else {
            bird.wingAngle = Math.sin(frameCount * 0.1) * 0.3;
        }
        
        // 检查地面碰撞
        if (bird.y + bird.radius >= CANVAS_HEIGHT - 80) {
            bird.y = CANVAS_HEIGHT - 80 - bird.radius;
            gameOver();
            return;
        }
        
        // 检查天花板碰撞
        if (bird.y - bird.radius <= 0) {
            bird.y = bird.radius;
            bird.velocity = 0;
        }
        
        // 生成管道
        if (frameCount % PIPE_SPAWN_RATE === 0) {
            var minHeight = 50;
            var maxHeight = CANVAS_HEIGHT - PIPE_GAP - minHeight - 80;
            var topHeight = Math.floor(Math.random() * (maxHeight - minHeight) + minHeight);
            pipes.push({
                x: CANVAS_WIDTH,
                topHeight: topHeight,
                passed: false
            });
        }
        
        // 更新管道
        for (var i = pipes.length - 1; i >= 0; i--) {
            var pipe = pipes[i];
            pipe.x -= PIPE_SPEED;
            
            // 计分
            if (!pipe.passed && pipe.x + PIPE_WIDTH < bird.x) {
                pipe.passed = true;
                score++;
                if (gameScore) gameScore.textContent = score;
                pipesPassed++;
                playScoreSound(); // 播放得分音效
            }
            
            // 碰撞检测
            if (bird.x + bird.radius > pipe.x && bird.x - bird.radius < pipe.x + PIPE_WIDTH) {
                if (bird.y - bird.radius < pipe.topHeight || bird.y + bird.radius > pipe.topHeight + PIPE_GAP) {
                    gameOver();
                    return;
                }
            }
            
            // 移除屏幕外管道
            if (pipe.x + PIPE_WIDTH < 0) {
                pipes.splice(i, 1);
            }
        }
    }
    
    // 绘制游戏
    function draw() {
        ctx.clearRect(0, 0, CANVAS_WIDTH, CANVAS_HEIGHT);
        
        drawBackground();
        
        // 绘制管道
        pipes.forEach(drawPipe);
        
        drawGround();
        drawBird();
    }
    
    // 游戏循环
    function gameLoop() {
        update();
        draw();
        animationId = requestAnimationFrame(gameLoop);
    }
    
    // 跳跃
    function jump() {
        // 首次跳跃时初始化音频上下文（需要用户交互）
        if (!audioCtx) {
            initAudio();
        }
        
        if (!isPlaying) {
            startGame();
            playJumpSound(); // 播放跳跃音效
            return;
        }
        if (isPaused) return;
        bird.velocity = JUMP_STRENGTH;
        playJumpSound(); // 播放跳跃音效
    }
    
    // 开始游戏（全局函数，供外部调用）
    window.startGame = function() {
        // 先取消之前的动画帧，防止多个循环同时运行
        cancelAnimationFrame(animationId);
        isPlaying = true;
        isPaused = false;
        gameStartTime = Date.now();
        
        // 启动背景音乐
        startBGM();
        
        if (startHint) startHint.classList.remove('show');
        
        resetGame();
        gameLoop();
    };
    
    // 重置游戏（全局函数）
    window.resetGame = function() {
        bird.y = 250;
        bird.velocity = 0;
        bird.rotation = 0;
        bird.wingAngle = 0;
        pipes = [];
        score = 0;
        pipesPassed = 0;
        frameCount = 0;
        if (gameScore) gameScore.textContent = '0';
    };
    
    // 格式化游戏时长
    function formatGameTime(ms) {
        var seconds = Math.floor(ms / 1000);
        var mins = Math.floor(seconds / 60);
        var secs = seconds % 60;
        return (mins < 10 ? '0' : '') + mins + ':' + (secs < 10 ? '0' : '') + secs;
    }
    
    // 历史最高分
    var highScore = 0;
    
    // 加载最高分
    function loadHighScore() {
        var saved = localStorage.getItem('flappyHighScore');
        if (saved) {
            highScore = parseInt(saved, 10);
        }
    }
    loadHighScore();
    
    // 保存最高分
    function saveHighScore() {
        if (score > highScore) {
            highScore = score;
            localStorage.setItem('flappyHighScore', highScore);
        }
    }
    
    // 游戏结束
    function gameOver() {
        isPlaying = false;
        cancelAnimationFrame(animationId);
        saveHighScore();
        playGameOverSound(); // 播放游戏结束音效
        stopBGM(); // 停止背景音乐
        
        // 触发gameOver事件，通知外部JS
        var gameOverEvent = new CustomEvent('gameOver', { detail: score });
        window.dispatchEvent(gameOverEvent);
    }
    
    // 暂停/继续
    window.togglePause = function() {
        if (!isPlaying) return;
        isPaused = !isPaused;
    };
    
    // 画布事件绑定
    canvas.addEventListener('mousedown', function(e) {
        e.preventDefault();
        jump();
    });
    
    canvas.addEventListener('touchstart', function(e) {
        e.preventDefault();
        jump();
    }, { passive: false });
    
    // 检查游戏结束弹窗是否显示
    function isGameOverModalVisible() {
        var modal = document.getElementById('gameOverModal');
        return modal && modal.classList.contains('show');
    }
    
    // 关闭游戏结束弹窗
    function closeGameOverModal() {
        var modal = document.getElementById('gameOverModal');
        if (modal && modal.classList.contains('show')) {
            // 触发关闭按钮点击事件
            var closeBtn = modal.querySelector('.modal-btn.secondary');
            if (closeBtn) closeBtn.click();
            return true;
        }
        return false;
    }
    
    // 检查当前是否在游戏页面
    function isInGamePage() {
        // 检查是否有游戏画布且画布在视口中可见
        var gameCanvas = document.getElementById('gameCanvas');
        if (!gameCanvas) return false;

        // 检查画布是否在视口内且尺寸正常（被渲染了）
        var rect = gameCanvas.getBoundingClientRect();
        return rect.width > 0 && rect.height > 0 && rect.top < window.innerHeight && rect.bottom > 0;
    }

    // 键盘控制
    document.addEventListener('keydown', function(e) {
        // 只有当在游戏页面时才响应键盘事件
        if (!isInGamePage()) {
            return;
        }

        // 如果游戏结束弹窗显示，空格键先关闭弹窗
        if ((e.code === 'Space' || e.code === 'ArrowUp' || e.code === 'Enter') && isGameOverModalVisible()) {
            e.preventDefault();
            closeGameOverModal();
            return;
        }

        // 只在画布可见时响应键盘
        if (!isPlaying && !startHint.classList.contains('show') && startHint.style.display === 'none') {
            return;
        }

        if (e.code === 'Space' || e.code === 'ArrowUp') {
            e.preventDefault();
            jump();
        } else if (e.code === 'KeyP') {
            e.preventDefault();
            togglePause();
        } else if (e.code === 'Enter') {
            e.preventDefault();
            if (!isPlaying && startHint && startHint.classList.contains('show') && startHint.style.display !== 'none') {
                startGame();
            }
        }
    });
    
    // 初始绘制
    draw();
