var check = document.getElementById('check');

check.addEventListener("animationend", AnimationHandler, false);

function AnimationHandler() {
    check.classList.remove('animated-check');
}

document.getElementById('configuration-submit').addEventListener('click', function (event) {
    check.classList.add('animated-check');
});