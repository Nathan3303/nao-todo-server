import express from 'express';
import { chat, todoChatter } from '@nao-todo-server/apis';

const exRouter: express.Router = express.Router();

exRouter.get('/ai/chat', async (req, res) => {
    await chat(req, res);
});

exRouter.post('/ai/todochatter', async (req, res) => {
    await todoChatter(req, res);
});

export default exRouter;
