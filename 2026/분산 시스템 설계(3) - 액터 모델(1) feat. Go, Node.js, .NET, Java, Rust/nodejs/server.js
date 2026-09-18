// server.js (CommonJS)
const express = require('express');
const app = express();

app.get('/', (req, res) => {
  res.send('hello');
});

app.listen(3000, () => console.log('listening')); // 내부적으로 콜백을 process.nextTick으로 등록
process.nextTick(() => console.log('tick'));
Promise.resolve().then(() => console.log('promise'));
console.log('end of script');