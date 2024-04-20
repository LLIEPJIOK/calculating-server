document.body.addEventListener('htmx:beforeRequest', function (event) {
    document.getElementById('spinner').style.display = 'block';
    document.getElementById('submit-button').style.display = 'none';
});

document.body.addEventListener('htmx:afterRequest', function (event) {
    document.getElementById('spinner').style.display = 'none';
    document.getElementById('submit-button').style.display = 'block';
});