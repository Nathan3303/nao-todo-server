import express from 'express';
import cors from 'cors';
import path from 'path';
import fs from 'fs';
import https from 'https';
import { fileURLToPath } from 'url';
import mongoose from 'mongoose';
import aiRoutes from './routes/ai';
import authRoutes from './routes/auth';
import projectRoutes from './routes/project';
import todoRoutes from './routes/todo';
import eventRoutes from './routes/event';
import tagRoutes from './routes/tag';
import userRoutes from './routes/user';
import commentRoutes from './routes/comment';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const uploadDir = path.join(__dirname, 'avatars');

const app = express();

app.use(cors());
app.use(express.json());
app.use(express.urlencoded({ extended: true }));

app.use('/statics/avatars', express.static(uploadDir, { maxAge: 31536000 }));
app.use('/api', aiRoutes);
app.use('/api', authRoutes);
app.use('/api', projectRoutes);
app.use('/api', todoRoutes);
app.use('/api', eventRoutes);
app.use('/api', tagRoutes);
app.use('/api', userRoutes);
app.use('/api', commentRoutes);
app.use('/', (_, res) => res.end('Hello World!') && void 0);

if (PROD) {
    mongoose
        .connect(`mongodb://172.18.0.3:27017/naotodo`)
        .then(() => {
            https
                .createServer(
                    {
                        key: fs.readFileSync(
                            path.join(process.cwd(), 'certs/privkey.pem')
                        ),
                        cert: fs.readFileSync(
                            path.join(process.cwd(), 'certs/fullchain.pem')
                        )
                    },
                    app
                )
                .listen(3002, () => {
                    console.log(
                        'NaoTodoServer(production) is running on port 3002'
                    );
                });
        });
} else {
    mongoose.connect(`mongodb://localhost:27017/naotodo`).then(() => {
        app.listen(3002, () => {
            console.log('NaoTodoServer(development) is running on port 3002');
        });
    });
}
