import express from 'express';
import cors from 'cors';
import path from 'path';
import fs from 'fs';
import https from 'https';
import mongoose from 'mongoose';
import authRoutes from './routes/auth';
import projectRoutes from './routes/project';
import todoRoutes from './routes/todo';
import eventRoutes from './routes/event';
import tagRoutes from './routes/tag';
import userRoutes from './routes/user';

const app = express();

app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.use('/api', authRoutes);
app.use('/api', projectRoutes);
app.use('/api', todoRoutes);
app.use('/api', eventRoutes);
app.use('/api', tagRoutes);
app.use('/api', userRoutes);

app.use('/', (_, res) => {
    res.end('Hello World!');
});

const keyPath = path.join(process.cwd(), 'certs/nathan33.xyz.key');
const certPath = path.join(process.cwd(), 'certs/nathan33.xyz.pem');
const mongodbUrl = `mongodb://${PROD ? '172.18.0.3' : 'localhost'}:27017/naotodo`;

if (PROD) {
    mongoose.connect(mongodbUrl).then(() => {
        https
            .createServer(
                {
                    key: fs.readFileSync(keyPath),
                    cert: fs.readFileSync(certPath)
                },
                app
            )
            .listen(3002, () => {
                console.log(
                    'NaoTodoServer(production) is running on port 3002'
                );
            });
    });
} else if (DEV) {
    mongoose.connect(mongodbUrl).then(() => {
        app.listen(3002, () => {
            console.log('NaoTodoServer(development) is running on port 3002');
        });
    });
} else {
    console.log('NaoTodoServer is not running');
}
