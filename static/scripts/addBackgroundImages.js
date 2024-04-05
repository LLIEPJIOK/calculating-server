document.addEventListener("DOMContentLoaded", function () {
    const container = document.getElementById('image-container');
    const width = document.documentElement.clientWidth;
    const height = document.documentElement.clientHeight;

    const row = Math.ceil(height / 70);
    const col = Math.ceil(width / 70);

    for (let i = 0; i < row; i++) {
        for (let j = 0; j < col; j++) {
            let div = document.createElement('div');
            div.classList.add('background-image');
            const randomNumber = Math.floor(Math.random() * 10) + 1;
            div.style.backgroundImage = `url(/static/resources/images/backgroundImages/image${randomNumber}.png)`;
            container.appendChild(div);
        }
    }
});