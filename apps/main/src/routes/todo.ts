import express from 'express';
import {
    createTodo,
    deleteTodo,
    updateTodo,
    updateTodos,
    getTodo,
    getTodos
} from '@nao-todo-server/apis';

const exRouter: express.Router = express.Router();

exRouter.get('/todo', async (req, res) => {
    await getTodo(req, res);
});

exRouter.get('/todos', async (req, res) => {
    await getTodos(req, res);
});

exRouter.post('/todo', async (req, res) => {
    await createTodo(req, res);
});

exRouter.put('/todo', async (req, res) => {
    await updateTodo(req, res);
});

exRouter.put('/todos', async (req, res) => {
    await updateTodos(req, res);
});

exRouter.delete('/todo', async (req, res) => {
    await deleteTodo(req, res);
});

export default exRouter;
