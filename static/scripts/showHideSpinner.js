document.body.addEventListener('htmx:afterRequest', function (event) {
    document.getElementById('spinner').style.opacity = 0;
});

document.body.addEventListener('htmx:beforeRequest', function (event) {
    document.getElementById('spinner').style.opacity = 1;
});