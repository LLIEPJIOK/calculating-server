var configurationForm = document.getElementById('configuration-form');

configurationForm.addEventListener('input', function (event) {
    if (event.target.tagName.toLowerCase() === 'input') {
        localStorage.setItem(event.target.name, event.target.value);
    }
});

window.addEventListener('load', function () {
    var inputs = configurationForm.getElementsByTagName('input');
    for (var i = 0; i < inputs.length; i++) {
        var storedValue = localStorage.getItem(inputs[i].name);
        if (storedValue) {
            inputs[i].value = storedValue;
        }
    }
});