/**
 * LEXIO - Authentication & Options Module
 * Flutter-inspired Client-Side Account Management
 */

document.addEventListener('DOMContentLoaded', function() {
    initAuth();
});

function initAuth() {
    const loginPage = document.getElementById('loginPage');
    const mainApp = document.getElementById('mainApp');
    const loginForm = document.getElementById('loginForm');
    const logoutBtn = document.getElementById('logoutBtn');
    const userNameDisplay = document.getElementById('userName');
    const userInitials = document.getElementById('userInitials');
    const userAvatar = document.getElementById('userAvatar');

    // Input fields
    const nameInput = document.getElementById('name');
    const usernameInput = document.getElementById('username');
    const emailInput = document.getElementById('email');
    const passwordInput = document.getElementById('password');
    const phoneInput = document.getElementById('phone');
    const avatarInput = document.getElementById('avatarInput');
    const avatarPickerBtn = document.getElementById('avatarPickerBtn');
    const loginAvatarPreview = document.getElementById('loginAvatarPreview');

    // Mode tabs & buttons
    const tabCreateAcc = document.getElementById('tabCreateAcc');
    const tabSignIn = document.getElementById('tabSignIn');
    const authSubmitBtn = document.getElementById('authSubmitBtn');
    const openOptionsBtn = document.getElementById('openOptionsBtn');
    const quickDemoBtn = document.getElementById('quickDemoBtn');
    const navGetStartedBtn = document.getElementById('navGetStartedBtn');

    // Options Modal elements
    const optionsModal = document.getElementById('optionsModal');
    const closeOptionsBtn = document.getElementById('closeOptionsBtn');
    const saveOptionsBtn = document.getElementById('saveOptionsBtn');
    const modalQuickDemoBtn = document.getElementById('modalQuickDemoBtn');

    // Fields that toggle between Create Account and Sign In
    const nameContainer = document.getElementById('nameContainer');
    const avatarRow = document.getElementById('avatarRow');

    let currentMode = 'create'; // 'create' or 'signin'
    let currentUser = null;
    let avatarDataUrl = '';

    function showSafeStatus(message, type) {
        const tone = type || 'error';
        if (typeof window.showStatus === 'function') {
            window.showStatus(message, tone);
            return;
        }

        let statusDiv = document.getElementById('status');
        if (!statusDiv) {
            statusDiv = document.createElement('div');
            statusDiv.id = 'status';
            statusDiv.className = 'status-indicator status-floating';
            document.body.appendChild(statusDiv);
        }

        statusDiv.textContent = message;
        statusDiv.style.display = 'inline-flex';
        statusDiv.className = 'status-indicator status-floating visible';

        setTimeout(function() {
            if (statusDiv) {
                statusDiv.classList.remove('visible');
            }
        }, 3200);
    }

    function validateEmail(email) {
        const re = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
        return re.test(email);
    }

    function getUserInitials(name) {
        if (!name) return 'U';
        return name
            .split(/\s+/)
            .filter(Boolean)
            .map(function(part) { return part.charAt(0); })
            .join('')
            .toUpperCase()
            .substring(0, 2) || 'U';
    }

    function addSafeActivity(message) {
        if (typeof window.addActivity === 'function') {
            window.addActivity(message);
        } else {
            console.log('Activity:', message);
        }
    }

    function paintAvatar(element, initialsElement, name, imageData, fallbackInitials) {
        if (!element) return;

        const initials = fallbackInitials || getUserInitials(name);
        const hue = (name || initials).split('').reduce(function(total, char) {
            return total + char.charCodeAt(0);
        }, 0) % 360;

        element.style.backgroundImage = '';
        element.style.backgroundSize = '';
        element.style.backgroundPosition = '';
        element.style.color = '#ffffff';
        element.classList.remove('has-photo');

        if (imageData) {
            element.style.background = '#0468D7';
            element.style.backgroundImage = 'url(' + imageData + ')';
            element.style.backgroundSize = 'cover';
            element.style.backgroundPosition = 'center';
            element.classList.add('has-photo');
            if (initialsElement) {
                initialsElement.textContent = initials;
                initialsElement.style.opacity = '0';
            } else {
                element.textContent = '';
            }
            return;
        }

        element.style.background = 'linear-gradient(135deg, hsl(' + hue + ', 80%, 48%), hsl(' + ((hue + 45) % 360) + ', 88%, 56%))';
        if (initialsElement) {
            initialsElement.textContent = initials;
            initialsElement.style.opacity = '1';
        } else {
            element.textContent = initials;
        }
    }

    function syncLoginAvatarPreview() {
        if (!loginAvatarPreview) return;
        const previewName = nameInput && nameInput.value.trim() ? nameInput.value.trim() : (usernameInput && usernameInput.value.trim() ? usernameInput.value.trim() : 'Alex Brooks');
        paintAvatar(loginAvatarPreview, null, previewName, avatarDataUrl, getUserInitials(previewName));
    }

    // Avatar upload handlers
    if (avatarPickerBtn && avatarInput) {
        avatarPickerBtn.addEventListener('click', function(e) {
            e.preventDefault();
            avatarInput.click();
        });
    }

    if (avatarInput) {
        avatarInput.addEventListener('change', function(event) {
            const file = event.target.files && event.target.files[0] ? event.target.files[0] : null;
            if (!file) {
                avatarDataUrl = '';
                syncLoginAvatarPreview();
                return;
            }

            if (!file.type || file.type.indexOf('image/') !== 0) {
                showSafeStatus('Please upload an image file (PNG, JPG, SVG, WebP)', 'error');
                avatarInput.value = '';
                return;
            }

            const reader = new FileReader();
            reader.onload = function(loadEvent) {
                avatarDataUrl = typeof loadEvent.target.result === 'string' ? loadEvent.target.result : '';
                syncLoginAvatarPreview();
                showSafeStatus('Display photo updated', 'success');
            };
            reader.readAsDataURL(file);
        });
    }

    if (nameInput) {
        nameInput.addEventListener('input', function() {
            if (!avatarDataUrl) {
                syncLoginAvatarPreview();
            }
        });
    }

    if (usernameInput) {
        usernameInput.addEventListener('input', function() {
            if (!avatarDataUrl && (!nameInput || !nameInput.value.trim())) {
                syncLoginAvatarPreview();
            }
        });
    }

    // Tab Mode Switcher: Create Account vs Sign In
    function setMode(mode) {
        currentMode = mode;
        if (mode === 'create') {
            if (tabCreateAcc) tabCreateAcc.classList.add('active');
            if (tabSignIn) tabSignIn.classList.remove('active');
            if (nameContainer) nameContainer.style.display = 'flex';
            if (avatarRow) avatarRow.style.display = 'flex';
            if (authSubmitBtn) {
                authSubmitBtn.innerHTML = '<span>Create Account &amp; Enter</span> <i class="fas fa-arrow-right"></i>';
            }
        } else {
            if (tabSignIn) tabSignIn.classList.add('active');
            if (tabCreateAcc) tabCreateAcc.classList.remove('active');
            if (nameContainer) nameContainer.style.display = 'none';
            if (avatarRow) avatarRow.style.display = 'none';
            if (authSubmitBtn) {
                authSubmitBtn.innerHTML = '<span>Sign In to Workspace</span> <i class="fas fa-arrow-right"></i>';
            }
        }
    }

    if (tabCreateAcc) {
        tabCreateAcc.addEventListener('click', function() { setMode('create'); });
    }

    if (tabSignIn) {
        tabSignIn.addEventListener('click', function() { setMode('signin'); });
    }

    // Quick navigation scroll to auth card
    if (navGetStartedBtn) {
        navGetStartedBtn.addEventListener('click', function() {
            const authCard = document.querySelector('.auth-card');
            if (authCard) {
                authCard.scrollIntoView({ behavior: 'smooth', block: 'center' });
            }
        });
    }

    // Workspace Options Modal Handlers
    function openOptions() {
        if (optionsModal) {
            optionsModal.classList.add('active');
        }
    }

    function closeOptions() {
        if (optionsModal) {
            optionsModal.classList.remove('active');
        }
    }

    if (openOptionsBtn) {
        openOptionsBtn.addEventListener('click', openOptions);
    }

    if (closeOptionsBtn) {
        closeOptionsBtn.addEventListener('click', closeOptions);
    }

    if (optionsModal) {
        optionsModal.addEventListener('click', function(e) {
            if (e.target === optionsModal) {
                closeOptions();
            }
        });
    }

    if (saveOptionsBtn) {
        saveOptionsBtn.addEventListener('click', function() {
            const optLang = document.getElementById('optDefaultLang');
            const optPace = document.getElementById('optDefaultPace');
            if (optLang && optLang.value) {
                localStorage.setItem('lexio_opt_lang', optLang.value);
            }
            if (optPace && optPace.value) {
                localStorage.setItem('lexio_opt_pace', optPace.value);
            }
            showSafeStatus('Workspace preferences saved', 'success');
            closeOptions();
        });
    }

    // Launch Workspace Helper
    function launchWorkspace(user) {
        currentUser = user;
        window.currentUser = currentUser;

        if (userNameDisplay) {
            userNameDisplay.textContent = user.name || user.username;
        }
        if (userInitials) {
            userInitials.textContent = getUserInitials(user.name || user.username);
        }
        paintAvatar(userAvatar, userInitials, user.name || user.username, user.avatar, getUserInitials(user.name || user.username));

        // Save session
        try {
            localStorage.setItem('lexio_current_user', JSON.stringify(currentUser));
        } catch (e) {
            // LocalStorage might be full or restricted
        }

        if (loginPage) {
            loginPage.style.display = 'none';
        }
        if (mainApp) {
            mainApp.style.display = 'flex';
        }

        addSafeActivity('User ' + (user.name || user.username) + ' entered workspace');
        showSafeStatus('Welcome, ' + (user.name || user.username) + '!', 'success');
    }

    // Instant Guest Demo Button Handlers
    function runQuickDemo() {
        closeOptions();
        const demoUser = {
            name: 'Liam Brooks',
            username: 'liam_b',
            email: 'liam.brooks@voxionlabs.io',
            phone: '+1 415 555 0188',
            avatar: ''
        };
        launchWorkspace(demoUser);
    }

    if (quickDemoBtn) {
        quickDemoBtn.addEventListener('click', runQuickDemo);
    }

    if (modalQuickDemoBtn) {
        modalQuickDemoBtn.addEventListener('click', runQuickDemo);
    }

    // Form Submit Handler
    if (loginForm) {
        loginForm.addEventListener('submit', function(e) {
            e.preventDefault();

            const name = nameInput ? nameInput.value.trim() : '';
            const username = usernameInput ? usernameInput.value.trim() : '';
            const email = emailInput ? emailInput.value.trim() : '';
            const password = passwordInput ? passwordInput.value.trim() : '';
            const phone = phoneInput && phoneInput.value.trim() ? phoneInput.value.trim() : '+1 555 0100';

            if (!email || !password) {
                showSafeStatus('Please enter your email and password', 'error');
                return;
            }

            if (!validateEmail(email)) {
                showSafeStatus('Please provide a valid email address', 'error');
                return;
            }

            if (password.length < 4) {
                showSafeStatus('Password must be at least 4 characters long', 'error');
                return;
            }

            if (currentMode === 'create') {
                if (!name) {
                    showSafeStatus('Please enter your full name', 'error');
                    return;
                }
                if (!username) {
                    showSafeStatus('Please pick a workspace username', 'error');
                    return;
                }

                // Register account locally
                const newAccount = {
                    name: name,
                    username: username,
                    email: email,
                    phone: phone,
                    avatar: avatarDataUrl
                };

                try {
                    let accounts = JSON.parse(localStorage.getItem('lexio_accounts') || '[]');
                    accounts = accounts.filter(function(a) { return a.email !== email && a.username !== username; });
                    accounts.push(newAccount);
                    localStorage.setItem('lexio_accounts', JSON.stringify(accounts));
                } catch (err) {
                    // Ignore quota errors
                }

                launchWorkspace(newAccount);
            } else {
                // Sign In mode
                let loadedUser = {
                    name: username || name || email.split('@')[0],
                    username: username || email.split('@')[0],
                    email: email,
                    phone: phone,
                    avatar: avatarDataUrl
                };

                try {
                    const accounts = JSON.parse(localStorage.getItem('lexio_accounts') || '[]');
                    const match = accounts.find(function(a) {
                        return (a.email && a.email.toLowerCase() === email.toLowerCase()) ||
                               (a.username && a.username.toLowerCase() === email.toLowerCase());
                    });
                    if (match) {
                        loadedUser = match;
                    }
                } catch (err) {}

                launchWorkspace(loadedUser);
            }
        });
    }

    // Logout Handler
    if (logoutBtn) {
        logoutBtn.addEventListener('click', function() {
            if (window.speechSynthesis && window.speechSynthesis.speaking) {
                window.speechSynthesis.cancel();
            }

            currentUser = null;
            window.currentUser = null;
            avatarDataUrl = '';

            try {
                localStorage.removeItem('lexio_current_user');
            } catch (e) {}

            if (loginForm) {
                loginForm.reset();
            }

            syncLoginAvatarPreview();
            paintAvatar(userAvatar, userInitials, 'User', '', 'U');

            if (mainApp) {
                mainApp.style.display = 'none';
            }
            if (loginPage) {
                loginPage.style.display = 'flex';
                window.scrollTo({ top: 0, behavior: 'smooth' });
            }

            addSafeActivity('User logged out');
        });
    }

    // Dev query param bypass (?dev)
    const urlParams = new URLSearchParams(window.location.search);
    if (urlParams.has('dev')) {
        launchWorkspace({
            name: 'Developer',
            username: 'developer',
            email: 'dev@local.lexio',
            phone: '+1 415 555 0100',
            avatar: ''
        });
    }

    syncLoginAvatarPreview();
}

console.log('Flutter-style Auth & Options module loaded');
