import express from 'express';
import {
    createProject,
    deleteProject,
    updateProject,
    getProject,
    getProjects
} from '@nao-todo-server/apis';

const exRouter: express.Router = express.Router();

exRouter.post('/project', async (req, res) => {
    await createProject(req, res);
});

exRouter.delete('/project', async (req, res) => {
    await deleteProject(req, res);
});

exRouter.put('/project', async (req, res) => {
    await updateProject(req, res);
});

exRouter.get('/project', async (req, res) => {
    await getProject(req, res);
});

exRouter.get('/projects', async (req, res) => {
    await getProjects(req, res);
});

export default exRouter;
