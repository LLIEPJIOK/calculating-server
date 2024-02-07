document.addEventListener('htmx:afterRequest', function (event) {
    document.getElementById('expression').value = '';
});