// web/static/js/pages/admin-setup.js
document.addEventListener('DOMContentLoaded', function () {
    const form = document.querySelector('form');

    // 비밀번호 일치 확인 (Alpine.js에서도 처리하지만 추가 안전장치)
    form.addEventListener('submit', function (event) {
        const password = document.getElementById('password').value;
        const passwordConfirm = document.getElementById('password_confirm').value;

        if (password !== passwordConfirm) {
            event.preventDefault();
            alert('비밀번호가 일치하지 않습니다.');
            return false;
        }

        // 복잡성 검증 - 최소 8자 이상, 하나 이상의 숫자, 하나 이상의 특수문자
        if (password.length < 8) {
            event.preventDefault();
            alert('비밀번호는 최소 8자 이상이어야 합니다.');
            return false;
        }

        const hasNumber = /\d/.test(password);
        const hasSpecial = /[!@#$%^&*(),.?":{}|<>]/.test(password);

        if (!hasNumber || !hasSpecial) {
            // 경고만 표시하고 계속 진행 가능
            if (!confirm('비밀번호에 숫자와 특수문자를 포함하는 것이 권장됩니다. 계속 진행하시겠습니까?')) {
                event.preventDefault();
                return false;
            }
        }

        return true;
    });
});