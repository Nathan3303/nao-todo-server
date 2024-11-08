import express from 'express';
import cors from 'cors';
// import jwt from 'jsonwebtoken';
// import { jwtCryptoSecret } from '@nao-todo-server/hooks/src/use-jwt/constants';
// import { useErrorResponseData } from '@nao-todo-server/hooks';
import authRoutes from './routes/auth';
// import todoRoutes from './routes/todo';
// import projectRoutes from './routes/project';
// import eventRoutes from './routes/event';
// import tagRoutes from './routes/tag';

const app = express();

app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));
app.use('/api', authRoutes);
// app.use(function (request, response, next) {
//     const { authorization } = request.headers;
//     if (!authorization) {
//         response.json(useErrorResponseData('Authorization header is missing.'));
//         return;
//     }
//     const token = authorization.replace('Bearer ', '');
//     jwt.verify(token, jwtCryptoSecret, (err, _) => {
//         if (err) {
//             response.json(useErrorResponseData('Invalid token.'));
//             return;
//         }
//     });
//     next();
// });
// app.use('/api', todoRoutes);
// app.use('/api', projectRoutes);
// app.use('/api', eventRoutes);
// app.use('/api', tagRoutes);
app.use('/', (_, res) => {
    res.end('Hello World!');
});

app.listen(3002, () => {
    console.log('NaoTodoServer(dev) is running on port 3002');
});
