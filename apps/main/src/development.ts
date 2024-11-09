import express from 'express';
import cors from 'cors';
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

app.listen(3002, () => {
    console.log('NaoTodoServer(dev) is running on port 3002');
});
