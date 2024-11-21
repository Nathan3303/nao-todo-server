import express from 'express';
import cors from 'cors';
import https from 'https';
import fs from 'fs';
import authRoutes from './routes/auth';
import projectRoutes from './routes/project';
import todoRoutes from './routes/todo';
import eventRoutes from './routes/event';
import tagRoutes from './routes/tag';

const app = express();

app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.use('/api', authRoutes);
app.use('/api', projectRoutes);
app.use('/api', todoRoutes);
app.use('/api', eventRoutes);
app.use('/api', tagRoutes);

app.use('/', (_, res) => {
    res.end('Hello World!');
});

if (PROD) {
    https
        .createServer(
            {
                key: fs.readFileSync('certs/nathan33.xyz.key'),
                cert: fs.readFileSync('certs/nathan33.xyz.pem')
            },
            app
        )
        .listen(3002, () => {
            console.log('NaoTodoServer(production) is running on port 3002');
        });
} else if (DEV) {
    app.listen(3002, () => {
        console.log('NaoTodoServer(development) is running on port 3002');
    });
} else {
    console.log('NaoTodoServer is not running');
}
