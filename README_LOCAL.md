# SmartDoc AI - Test Local avec Volume

Ce guide explique comment tester l'API SmartDoc AI en local avec un volume de stockage persistant.

## 🚀 Démarrage Rapide

### 1. Lancer l'API avec Docker

```bash
# Construire et démarrer l'API
docker-compose up --build

# Ou en arrière-plan
docker-compose up -d --build
```

### 2. Vérifier que l'API fonctionne

```bash
# Test de santé
curl http://localhost:8080/health

# Swagger UI
open http://localhost:8080/swagger/
```

## 📁 Volume Local

L'API utilise un volume local pour stocker les fichiers uploadés :

- **Dossier local** : `./data/` (dans le projet)
- **Dossier dans le container** : `/app/data/`
- **Structure** : `./data/users/{userID}/documents/{docID}.{extension}`

### Exemple de structure des fichiers

```
data/
├── users/
│   ├── user123/
│   │   └── documents/
│   │       ├── doc_1234567890.pdf
│   │       └── doc_1234567891.jpg
│   └── user456/
│       └── documents/
│           └── doc_1234567892.png
```

## 🔧 Configuration

### Variables d'environnement par défaut

L'API fonctionne en mode local avec ces paramètres par défaut :

```yaml
STORAGE_TYPE=local
LOCAL_STORAGE_PATH=/app/data
ENV=development
LOG_LEVEL=info
```

### Services utilisés en local

- **Stockage** : Fichiers locaux (pas de Firebase Storage)
- **Base de données** : Mock (pas de Firestore)
- **OCR** : Mock (pas de service externe)
- **IA Summary** : Mock (pas de service externe)

## 🧪 Tests

### 1. Upload de document

```bash
# Avec curl
curl -X POST http://localhost:8080/documents \
  -H "Authorization: Bearer YOUR_FIREBASE_TOKEN" \
  -F "file=@/path/to/your/document.pdf"

# Avec Postman
POST http://localhost:8080/documents
Headers: Authorization: Bearer YOUR_FIREBASE_TOKEN
Body: form-data
  - file: [sélectionner un fichier]
```

### 2. Lister les documents

```bash
curl -X GET http://localhost:8080/documents \
  -H "Authorization: Bearer YOUR_FIREBASE_TOKEN"
```

### 3. Obtenir un document

```bash
curl -X GET http://localhost:8080/documents/{docID} \
  -H "Authorization: Bearer YOUR_FIREBASE_TOKEN"
```

### 4. Déclencher l'OCR

```bash
curl -X POST http://localhost:8080/documents/{docID}/ocr \
  -H "Authorization: Bearer YOUR_FIREBASE_TOKEN"
```

### 5. Générer un résumé

```bash
curl -X POST http://localhost:8080/documents/{docID}/summary \
  -H "Authorization: Bearer YOUR_FIREBASE_TOKEN"
```

## 🔐 Authentification

Pour tester l'API, tu as besoin d'un token Firebase valide :

1. **Option 1** : Utiliser Firebase Auth dans ton app
2. **Option 2** : Créer un token de test avec Firebase Admin SDK
3. **Option 3** : Modifier temporairement le middleware d'auth pour accepter un token de test

### Token de test (pour développement uniquement)

```bash
# Créer un token de test
curl -X POST http://localhost:8080/auth/test-token \
  -H "Content-Type: application/json" \
  -d '{"user_id": "test_user_123"}'
```

## 📊 Monitoring

### Logs

```bash
# Voir les logs en temps réel
docker-compose logs -f smartdoc-api

# Voir les logs d'un service spécifique
docker-compose logs smartdoc-api
```

### Métriques

```bash
# Santé de l'API
curl http://localhost:8080/health

# Informations système
curl http://localhost:8080/metrics
```

## 🛠️ Développement

### Modifier le code

1. Modifie les fichiers Go
2. Reconstruis l'image : `docker-compose build`
3. Redémarre : `docker-compose up`

### Debug

```bash
# Entrer dans le container
docker-compose exec smartdoc-api sh

# Voir les fichiers uploadés
ls -la /app/data/users/

# Vérifier les logs
tail -f /var/log/app.log
```

## 🧹 Nettoyage

### Supprimer les données

```bash
# Supprimer le volume local
rm -rf ./data/

# Ou supprimer seulement les fichiers d'un utilisateur
rm -rf ./data/users/user123/
```

### Arrêter l'API

```bash
# Arrêter les services
docker-compose down

# Arrêter et supprimer les volumes
docker-compose down -v
```

## 🔄 Migration vers Firebase

Quand tu veux utiliser Firebase en production :

1. Configure les variables d'environnement Firebase
2. Change `STORAGE_TYPE` de `local` à `firebase`
3. Redémarre l'API

```bash
# Variables d'environnement pour Firebase
export FIREBASE_PROJECT_ID=your-project-id
export FIREBASE_SERVICE_ACCOUNT_KEY='{"type":"service_account",...}'
export FIREBASE_STORAGE_BUCKET=your-bucket-name
export STORAGE_TYPE=firebase

# Redémarrer
docker-compose down
docker-compose up --build
```

## 🐛 Dépannage

### Problèmes courants

1. **Port déjà utilisé** : Change le port dans `docker-compose.yml`
2. **Permissions** : Vérifie les permissions du dossier `./data/`
3. **Token invalide** : Vérifie ton token Firebase
4. **Fichier trop gros** : Augmente la limite dans le serveur

### Logs d'erreur

```bash
# Voir les erreurs
docker-compose logs smartdoc-api | grep ERROR

# Voir les warnings
docker-compose logs smartdoc-api | grep WARN
```

## 📝 Notes

- Les fichiers sont persistants entre les redémarrages
- L'API fonctionne sans connexion internet
- Parfait pour le développement et les tests
- Les données sont isolées par utilisateur 