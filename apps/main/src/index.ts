import express from 'express';
import cors from 'cors';
import https from 'https';
import fs from 'fs';
import authRoutes from './routes/auth';
import todoRoutes from './routes/todo';
import projectRoutes from './routes/project';
import tagRoutes from './routes/tag';
import eventRoutes from './routes/event';
import { useErrorResponseData } from '@nao-todo-server/hooks';
import { jwtCryptoSecret } from '@nao-todo-server/hooks/src/use-jwt/constants';

const app = express();

app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));
app.use('/api', authRoutes);
app.use(function (request, response, next) {
    const { authorization } = request.headers;
    if (!authorization) {
        response.json(useErrorResponseData('Authorization header is missing.'));
        return;
    }
    const token = authorization.replace('Bearer ', '');
    next();
});
app.use('/api', todoRoutes);
app.use('/api', projectRoutes);
app.use('/api', tagRoutes);
app.use('/api', eventRoutes);
app.use('/', (_, res) => {
    res.end('Hello World!');
});

https
    .createServer(
        {
            key: fs.readFileSync('certs/nathan33.xyz.key'),
            cert: fs.readFileSync('certs/nathan33.xyz.pem')
        },
        app
    )
    .listen(3002, () => {
        console.log('NaoTodoServer(prod) is running on port 3002');
    });
