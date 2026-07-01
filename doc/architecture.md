# Arquitetura

## Fluxo de dados
1. Usuário configura filtros via Web/API
2. Worker executa a cada 15min (Cloud Scheduler)
3. Worker consulta Skyscanner (ou cache Firestore)
4. Aplica filtros (pkg/filters)
5. Notifica via SendGrid (email) e FCM (push)

## Limites GCP Always Free
- Cloud Scheduler: 3 jobs gratuitos
- Cloud Run: 2M invocações/mês
- Firestore: 1GB / 50k leituras diárias
- Cloud Storage: 5GB (para estáticos, se não usar Firebase Hosting)
