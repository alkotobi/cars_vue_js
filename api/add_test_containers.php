<?php
header('Content-Type: application/json');
require_once 'config.php';

try {
    $pdo = new PDO(
        "mysql:host=$db_host;dbname=$db_name;charset=utf8",
        $db_user,
        $db_pass
    );
    $pdo->setAttribute(PDO::ATTR_ERRMODE, PDO::ERRMODE_EXCEPTION);

    // Add some test containers
    $containers = [
        ['name' => 'Container A'],
        ['name' => 'Container B'],
        ['name' => 'Container C'],
        ['name' => 'Container D'],
        ['name' => 'Container E']
    ];

    foreach ($containers as $container) {
        $stmt = $pdo->prepare("INSERT IGNORE INTO containers (name) VALUES (?)");
        $stmt->execute([$container['name']]);
    }

    // Get the first loading record
    $stmt = $pdo->query("SELECT id FROM loading ORDER BY id DESC LIMIT 1");
    $loadingRecord = $stmt->fetch(PDO::FETCH_ASSOC);

    if ($loadingRecord) {
        $loadingId = $loadingRecord['id'];
        
        // Add some test loaded containers
        $loadedContainers = [
            [
                'id_loading' => $loadingId,
                'id_container' => 1,
                'ref_container' => 'REF001',
                'date_departed' => '2025-01-15',
                'date_loaded' => '2025-01-10',
                'note' => 'Test container 1'
            ],
            [
                'id_loading' => $loadingId,
                'id_container' => 2,
                'ref_container' => 'REF002',
                'date_departed' => '2025-01-16',
                'date_loaded' => '2025-01-11',
                'note' => 'Test container 2'
            ],
            [
                'id_loading' => $loadingId,
                'id_container' => 3,
                'ref_container' => 'REF003',
                'date_departed' => '2025-01-17',
                'date_loaded' => '2025-01-12',
                'note' => 'Test container 3'
            ]
        ];

        foreach ($loadedContainers as $loadedContainer) {
            $stmt = $pdo->prepare("
                INSERT IGNORE INTO loaded_containers 
                (id_loading, id_container, ref_container, date_departed, date_loaded, note) 
                VALUES (?, ?, ?, ?, ?, ?)
            ");
            $stmt->execute([
                $loadedContainer['id_loading'],
                $loadedContainer['id_container'],
                $loadedContainer['ref_container'],
                $loadedContainer['date_departed'],
                $loadedContainer['date_loaded'],
                $loadedContainer['note']
            ]);
        }

        echo json_encode([
            'success' => true,
            'message' => 'Test containers added successfully',
            'loading_id' => $loadingId,
            'containers_added' => count($containers),
            'loaded_containers_added' => count($loadedContainers)
        ]);
    } else {
        echo json_encode([
            'success' => false,
            'message' => 'No loading records found. Please add a loading record first.'
        ]);
    }

} catch(PDOException $e) {
    echo json_encode([
        'success' => false,
        'error' => $e->getMessage()
    ]);
}
?> 