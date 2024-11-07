import express from 'express';
import { signin, signup, checkin, signout } from '@nao-todo-server/apis';
// import type { Request, Response } from 'express';

const exRouter: express.Router = express.Router();

exRouter.post('/signin', async (req, res) => {
    await signin(req, res);
});

exRouter.post('/signup', async (req, res) => {
    await signup(req, res);
});

exRouter.get('/checkin', async (req, res) => {
    await checkin(req, res);
});

exRouter.delete('/signout', async (req, res) => {
    await signout(req, res);
});

export default exRouter;
