var ws;
var username;

function connect() {
    username = document.getElementById('username').value;
    ws = new WebSocket("ws://localhost:8080/ws?username=" + username);

    ws.onopen = function() {
        // Fetch and display online users once the WebSocket connection is established
        getUsers();
    };

    ws.onmessage = function(event) {
        var data = JSON.parse(event.data);
        if (data.type === "userList") {
            updateUsers(data.users);
        } else if (data.type === "message") {
            var messages = document.getElementById('messages');
            var message = document.createElement('li');
            message.textContent = data.sender + ": " + data.message;
            messages.appendChild(message);
        }
    };
}

function sendMessage() {
    var recipient = document.getElementById('recipient').value;
    var message = document.getElementById('message').value;
    ws.send(JSON.stringify({ type: "message", sender: username, recipient: recipient, message: message }));
}

function getUsers() {
    fetch('/users')
        .then(response => response.json())
        .then(data => {
            updateUsers(data.users);
        });
}

function updateUsers(users) {
    var usersList = document.getElementById('users');
    usersList.innerHTML = '';
    users.forEach(user => {
        var li = document.createElement('li');
        li.textContent = user;
        usersList.appendChild(li);
    });
}