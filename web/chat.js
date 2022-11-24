var scheme = window.location.protocol == "https:" ? 'wss://' : 'ws://';
var webSocketURL = scheme
  + 'localhost'
  + (location.port ? ':' + location.port : '')
  + '/chat';
var webSocket = new WebSocket(webSocketURL);

webSocket.onmessage = function (event) {
  addChat(event.data)
}

document.getElementById("chat-submit").onsubmit = function(event) {
  event.preventDefault();
  var inputElement = document.getElementById("chat-input");
  text = inputElement.value;
  inputElement.value = "";

  if(webSocket.readyState === webSocket.OPEN) {
    webSocket.send(text);
  }
}

function addChat(chat) {
  chatObj = JSON.parse(chat)

  var chatElement = document.createElement('div')
  chatElement.innerHTML = `<strong>${chatObj.userId}</strong>: ${chatObj.message}`

  document.getElementById("chat").append(chatElement);
}
